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
import { placeOrder } from "../actions/order-action";

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
		commonRequestParameters: {
			headers: {
				Authorization: jwt.jwtToken,
			},
		},
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

	// todo counter metrics/ end state
	// also based on user profile
}
