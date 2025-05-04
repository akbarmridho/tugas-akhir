import { Options } from "k6/options";
import { forgeJwt } from "../utils/jwt";
import {
	Order,
	TicketBackendServiceClient,
} from "../client/ticket/ticketBackendService";
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

export default function test() {
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
		return;
	}

	const state = getProfileState(profile, event);
	let order: Order | null = null;

	// place order phase
	while (state.currentOrderAttempt < state.maxOrderAttempt) {
		const orderItems = getSeat(ticketService, state);

		if (orderItems === null) {
			if (state.currentBrowseAttempt >= state.maxBrowseAttempt) {
				break;
			} else {
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
			return;
		} else if (orderResponse.state === "Retry") {
			state.currentOrderAttempt++;
			continue;
		}

		if (orderResponse.state !== "Ok" || orderResponse.order === null) {
			logger.error("Response should be ok and order is not null");
			return;
		}

		// continue payment
		order = orderResponse.order as Order;
	}

	// pay invoice
	// retry until 3 time if fail
	let tries = 0;
	let invoiceOk = false;
	const shouldSuccess = Math.random() <= 0.9;

	while (tries < 3) {
		const invoice = payOrder(paymentService, order as Order, shouldSuccess);

		if (invoice === null) {
			tries++;
			continue;
		}

		invoiceOk = true;
		break;
	}

	if (!invoiceOk) {
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
			}
		} else {
			if (newOrder.status === "success") {
				// not desirable
				logger.error("Expected order to be failed but received success");
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
		return;
	}

	// get issued tickets
	if (!shouldSuccess) {
		// this is end state for failed payment
		return;
	}

	tries = 0;

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
		break;
	}

	// end state

	// todo counter metrics/ end state
	// also based on user profile
}
