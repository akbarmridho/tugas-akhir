import { Options, Scenario } from "k6/options";
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

const { SCENARIO, DB_VARIANT, FC, RUN_ID, VARIANT } = __ENV;

const scenarioName = `${VARIANT}_${SCENARIO}`;

const generateScenario = (): Scenario => {
	if (VARIANT === "debug") {
		return {
			executor: "shared-iterations",
			vus: 2,
			iterations: 10,
			maxDuration: "5m",
		};
	} else if (VARIANT === "smoke") {
		return {
			executor: "constant-vus",
			vus: 50,
			duration: "5m",
		};
	} else if (VARIANT === "smokey") {
		return {
			executor: "constant-vus",
			vus: 200,
			duration: "10m",
		};
	} else if (VARIANT === "sim-1") {
		return {
			executor: "ramping-arrival-rate",
			preAllocatedVUs: 25000,
			stages: [
				{ target: 3964, duration: "30s" },
				{ target: 11850, duration: "30s" },
				{ target: 11500, duration: "30s" },
				{ target: 8329, duration: "30s" },
				{ target: 5280, duration: "30s" },
				{ target: 3244, duration: "30s" },
				{ target: 2100, duration: "30s" },
				{ target: 1300, duration: "30s" },
				{ target: 825, duration: "30s" },
				{ target: 515, duration: "30s" },
				{ target: 354, duration: "30s" },
				{ target: 235, duration: "30s" },
				{ target: 146, duration: "30s" },
				{ target: 115, duration: "30s" },
				{ target: 74, duration: "30s" },
				{ target: 55, duration: "30s" },
				{ target: 35, duration: "30s" },
				{ target: 18, duration: "30s" },
				{ target: 13, duration: "30s" },
				{ target: 9, duration: "30s" },
			],
		};
	} else if (VARIANT === "sim-2") {
		return {
			executor: "ramping-arrival-rate",
			preAllocatedVUs: 50000,
			stages: [
				{ target: 6606, duration: "30s" },
				{ target: 23664, duration: "30s" },
				{ target: 23777, duration: "30s" },
				{ target: 16826, duration: "30s" },
				{ target: 10861, duration: "30s" },
				{ target: 6789, duration: "30s" },
				{ target: 4138, duration: "30s" },
				{ target: 2559, duration: "30s" },
				{ target: 1622, duration: "30s" },
				{ target: 1069, duration: "30s" },
				{ target: 682, duration: "30s" },
				{ target: 446, duration: "30s" },
				{ target: 283, duration: "30s" },
				{ target: 193, duration: "30s" },
				{ target: 122, duration: "30s" },
				{ target: 102, duration: "30s" },
				{ target: 71, duration: "30s" },
				{ target: 61, duration: "30s" },
				{ target: 34, duration: "30s" },
				{ target: 22, duration: "30s" },
			],
		};
	} else if (VARIANT === "sim-test") {
		return {
			executor: "ramping-arrival-rate",
			preAllocatedVUs: 5000,
			stages: [
				{ target: 1118, duration: "30s" },
				{ target: 2918, duration: "30s" },
				{ target: 2696, duration: "30s" },
				{ target: 1955, duration: "30s" },
				{ target: 1210, duration: "30s" },
				{ target: 726, duration: "30s" },
				{ target: 507, duration: "30s" },
				{ target: 304, duration: "30s" },
				{ target: 168, duration: "30s" },
				{ target: 135, duration: "30s" },
				{ target: 88, duration: "30s" },
				{ target: 53, duration: "30s" },
				{ target: 40, duration: "30s" },
				{ target: 28, duration: "30s" },
				{ target: 22, duration: "30s" },
				{ target: 9, duration: "30s" },
				{ target: 4, duration: "30s" },
				{ target: 5, duration: "30s" },
				{ target: 5, duration: "30s" },
				{ target: 2, duration: "30s" },
			],
		};
	} else if (VARIANT === "stress-1") {
		return {
			executor: "shared-iterations",
			vus: 20000,
			iterations: 500000,
			maxDuration: "15m",
		};
	} else if (VARIANT === "stress-2") {
		return {
			executor: "shared-iterations",
			vus: 40000,
			iterations: 1000000,
			maxDuration: "15m",
		};
	} else if (VARIANT === "stress-test") {
		return {
			executor: "shared-iterations",
			vus: 4000,
			iterations: 12000,
			maxDuration: "15m",
		};
	}

	throw new Error("Invalid variant");
};

export const options: Options = {
	insecureSkipTLSVerify: true,
	tags: {
		scenario: SCENARIO,
		dbvariant: DB_VARIANT,
		fc: FC,
		testid: RUN_ID,
		variant: VARIANT,
	},
	scenarios: {
		[scenarioName]: generateScenario(),
	},
	hosts: {
		"payment.tugas-akhir.local": __ENV.HOST_FORWARD || "127.0.0.1",
		"ticket.tugas-akhir.local": __ENV.HOST_FORWARD || "127.0.0.1",
	},
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
	NOT_SET: "NOT_SET",
	UNKNOWN: "UNKNOWN_EXIT", // Fallback
};

export default function test() {
	const tags = vu.metrics.tags;

	const submitMetric = () => {
		const result: Record<string, string> = {};
		for (const key of Object.keys(tags)) {
			const tagVal = `${tags[key]}`;
			if (!tagVal) {
				continue;
			}
			result[key] = tagVal;
		}

		vuEndStateCounter.add(1, result);
	};

	tags.state = END_STATES.NOT_SET;

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
					Authorization: `Bearer ${jwt.jwtToken}`,
				},
			},
		});

		const event = getEvent(ticketService);

		if (!event) {
			// error
			// should not reach this
			logger.error("Event not found");
			tags.state = END_STATES.FAIL_EVENT_NOT_FOUND;
			submitMetric();
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
					submitMetric();
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
				submitMetric();
				return;
			} else if (orderResponse.state === "Retry") {
				state.currentOrderAttempt++;
				sleep(randomIntBetween(1, 5));
				continue;
			}

			if (orderResponse.state !== "Ok" || orderResponse.order === null) {
				logger.error("Response should be ok and order is not null");
				tags.state = END_STATES.UNKNOWN;
				submitMetric();
				return;
			}

			// continue payment
			order = orderResponse.order;
			break;
		}

		if (!order) {
			logger.error("Max order placement attempt reached");
			tags.state = END_STATES.FAIL_ORDER_ATTEMPTS_EXCEEDED;
			submitMetric();
			return;
		}

		// pay invoice
		// retry until 3 time if fail
		let tries = 0;
		let invoiceOk = false;
		const shouldSuccess = Math.random() <= 0.95;

		tags.profile_payment_success = shouldSuccess;

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
			submitMetric();
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
					submitMetric();
					return;
				}
			} else {
				if (newOrder.status === "success") {
					// not desirable
					logger.error("Expected order to be failed but received success");
					tags.state = END_STATES.ERROR_UNEXPECTED_ORDER_STATE;
					submitMetric();
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
			submitMetric();
			return;
		}

		// get issued tickets
		if (!shouldSuccess) {
			// this is end state for failed payment
			tags.state = END_STATES.SUCCESS_PAYMENT_FAILED_EXPECTED;
			submitMetric();
			return;
		}

		tries = 0;
		let ticketIssued = false;

		while (tries < 3) {
			const issuedTicketResponse = getIssuedTickets(ticketService, order!.id);

			if (issuedTicketResponse === null) {
				sleep(randomIntBetween(5, 15));
				tries++;
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
			submitMetric();
			return; // Exit
		}

		// --- Final Success State ---
		tags.state = END_STATES.SUCCESS_PURCHASED;
		submitMetric();
	} catch (e) {
		logger.error(`Unhandled exception in VU script: ${e}`);
		tags.state = END_STATES.UNKNOWN;
		submitMetric();
	}
}
