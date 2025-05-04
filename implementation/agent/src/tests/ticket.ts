import { Options } from "k6/options";
import { forgeJwt } from "../utils/jwt";
import {
	Order,
	TicketBackendServiceClient,
} from "../client/ticket/ticketBackendService";
import { vu } from "k6/execution";
import { Counter } from "k6/metrics";
import { PaymentServiceClient } from "../client/payment/paymentService";
import { getEvent } from "../actions/event";
import { getProfileState, getUserProfile } from "../utils/profile";
import { logger } from "../utils/logger";
import { getSeat } from "../actions/seats";
import { payOrder, placeOrder } from "../actions/order-action";
import { check, sleep } from "k6";
import { randomIntBetween } from "https://jslib.k6.io/k6-utils/1.2.0/index.js";
import { getIssuedTickets, getOrder } from "../actions/order-query";

export const options: Options = {
	insecureSkipTLSVerify: true,
	// todo pass as CLI parameter
	// tags: {
	// 	scenario: "sf-2",
	// 	dbvariant: "postgres",
	// 	fc: "nofc",
	// },
};

const vuEndStateCounter = new Counter("vu_end_state");

// Define possible end states as constants for consistency
const END_STATES = {
	SUCCESS_PURCHASED: "SUCCESS_PURCHASED",
	SUCCESS_PAYMENT_FAILED_EXPECTED: "SUCCESS_PAYMENT_FAILED_EXPECTED",
	FAIL_EVENT_NOT_FOUND: "FAIL_EVENT_NOT_FOUND",
	FAIL_SEAT_NOT_FOUND: "FAIL_SEAT_NOT_FOUND",
	FAIL_ORDER_BAD_REQUEST: "FAIL_ORDER_BAD_REQUEST",
	FAIL_ORDER_ATTEMPTS_EXCEEDED: "FAIL_ORDER_ATTEMPTS_EXCEEDED",
	FAIL_PAYMENT_SYSTEM_ERROR: "FAIL_PAYMENT_SYSTEM_ERROR",
	FAIL_ORDER_VERIFICATION: "FAIL_ORDER_VERIFICATION",
	FAIL_TICKET_ISSUANCE: "FAIL_TICKET_ISSUANCE",
	ERROR_UNEXPECTED_ORDER_STATE: "ERROR_UNEXPECTED_ORDER_STATE",
	UNKNOWN: "UNKNOWN_EXIT", // Fallback
};

export default function test() {
	const tags = vu.metrics.tags;

	tags.state = END_STATES.UNKNOWN;

	try {
		const jwt = forgeJwt();
		const profile = getUserProfile();

		const paymentService = new PaymentServiceClient({
			baseUrl: "https://payment.tugas-akhir.local",
		});

		const ticketService = new TicketBackendServiceClient({
			baseUrl: "https://ticket.tugas-akhir.local",
			commonRequestParameters: {
				headers: {
					Authorization: jwt.jwtToken,
				},
			},
		});

		const event = getEvent(ticketService);

		if (!event) {
			// error
			// should not reach this
			logger.error("Event not found");
			tags.state = END_STATES.FAIL_EVENT_NOT_FOUND;
			return;
		}

		const state = getProfileState(profile, event);

		tags.profile_tier = profile.seatingTier;
		tags.profile_day_preference = profile.dayPreference;
		tags.profile_quantity = profile.quantity;
		tags.profile_persistence = profile.persistence;
		tags.profile_fallback_type = state.fallbackType;

		let order: Order | null = null;

		// place order phase
		while (state.currentOrderAttempt < state.maxOrderAttempt) {
			const orderItems = getSeat(ticketService, state);

			if (orderItems === null) {
				if (state.currentBrowseAttempt >= state.maxBrowseAttempt) {
					tags.state = END_STATES.FAIL_SEAT_NOT_FOUND;
					break;
				} else {
					sleep(randomIntBetween(2, 8));
					continue;
				}
			}

			const orderResponse = placeOrder(
				ticketService,
				state,
				orderItems.items,
				orderItems.sale,
			);

			if (orderResponse.state === "Bad Request") {
				// should not reach this
				// error already logger
				logger.error("Order placement returned Bad Request");
				tags.state = END_STATES.FAIL_ORDER_BAD_REQUEST;
				return;
			} else if (orderResponse.state === "Retry") {
				state.currentOrderAttempt++;
				sleep(randomIntBetween(1, 5));
				continue;
			}

			if (orderResponse.state !== "Ok" || orderResponse.order === null) {
				logger.error("Response should be ok and order is not null");
				tags.state = END_STATES.UNKNOWN;
				return;
			}

			// continue payment
			order = orderResponse.order;
			break;
		}

		if (!order) {
			logger.error("Max order placement attempt reached");
			tags.state = END_STATES.FAIL_ORDER_ATTEMPTS_EXCEEDED;
			return;
		}

		// pay invoice
		// retry until 3 time if fail
		let tries = 0;
		let invoiceOk = false;
		const shouldSuccess = Math.random() <= 0.9;

		sleep(randomIntBetween(1, 3));

		while (tries < 3) {
			const invoice = payOrder(paymentService, order as Order, shouldSuccess);

			if (invoice === null) {
				tries++;
				sleep(randomIntBetween(1, 3));
				continue;
			}

			invoiceOk = true;
			break;
		}

		if (!invoiceOk) {
			logger.error("Failed to pay invoice after retries");
			tags.state = END_STATES.FAIL_PAYMENT_SYSTEM_ERROR;
			return;
		}

		// wait for the invoice payment to be completed
		sleep(randomIntBetween(5, 15));

		tries = 0;
		let orderConfirmend = false;

		while (tries < 3) {
			const newOrder = getOrder(ticketService, (order as Order).id);

			if (newOrder === null) {
				// retry
				tries++;
				sleep(randomIntBetween(5, 15));
				continue;
			}

			if (shouldSuccess) {
				if (newOrder.status === "success") {
					// reached desired end state
					orderConfirmend = true;
					break;
				} else if (newOrder.status === "waiting-for-payment") {
					// retry
					tries++;
					sleep(randomIntBetween(5, 15));
				} else if (newOrder.status === "failed") {
					// not desirable
					logger.error("Expected order to be success but received failed");
					tags.state = END_STATES.ERROR_UNEXPECTED_ORDER_STATE;
					return;
				}
			} else {
				if (newOrder.status === "success") {
					// not desirable
					logger.error("Expected order to be failed but received success");
					tags.state = END_STATES.ERROR_UNEXPECTED_ORDER_STATE;
					return;
				} else if (newOrder.status === "waiting-for-payment") {
					// retry
					tries++;
					sleep(randomIntBetween(5, 15));
				} else if (newOrder.status === "failed") {
					// reached desired end state
					orderConfirmend = true;
					break;
				}
			}
		}

		if (!orderConfirmend) {
			logger.error("Cannot verify the order status");
			tags.state = END_STATES.FAIL_ORDER_VERIFICATION;
			return;
		}

		// get issued tickets
		if (!shouldSuccess) {
			// this is end state for failed payment
			tags.state = END_STATES.SUCCESS_PAYMENT_FAILED_EXPECTED;
			return;
		}

		tries = 0;
		let ticketIssued = false;

		while (tries < 3) {
			const issuedTicketResponse = getIssuedTickets(ticketService, order!.id);

			if (issuedTicketResponse === null) {
				sleep(randomIntBetween(5, 15));
				continue;
			}

			const published = check(issuedTicketResponse, {
				"published tickets equals to purchased tickets": (t) =>
					t.length === state.ticketCount,
			});

			if (!published) {
				logger.error("Published tickets does not equal to purchased tickets");
			}

			ticketIssued = true;
			break;
		}

		if (!ticketIssued) {
			logger.error("Failed to get issued tickets after retries.");
			tags.state = END_STATES.FAIL_TICKET_ISSUANCE;
			return; // Exit
		}

		// --- Final Success State ---
		tags.state = END_STATES.SUCCESS_PURCHASED;
	} catch (e) {
		logger.error(`Unhandled exception in VU script: ${e}`);
		tags.state = END_STATES.UNKNOWN;
	} finally {
		const result: Record<string, string> = {};

		for (const key of Object.keys(tags)) {
			result[key] = `${tags[key]}`;
		}
		vuEndStateCounter.add(1, result);
	}
}
