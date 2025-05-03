import { group, sleep } from "k6";
import {
	Order,
	OrderItemDto,
	TicketBackendServiceClient,
	TicketSale,
} from "../client/ticket/ticketBackendService";
import { ProfileState } from "../utils/profile";
import { v4 as uuidv4 } from "https://jslib.k6.io/uuid/1.0.0/index.js";
import { randomIntBetween } from "https://jslib.k6.io/k6-utils/1.2.0/index.js";
import { logger } from "../utils/logger";
import { PaymentServiceClient } from "../client/payment/paymentService";

interface PlaceOrderResponse {
	state: "Ok" | "Bad Request" | "Retry";
	order: Order | null;
}

export const placeOrder = (
	ticketService: TicketBackendServiceClient,
	state: ProfileState,
	items: OrderItemDto[],
	sale: TicketSale,
): PlaceOrderResponse => {
	const payload = {
		eventId: state.saleSelection[0].eventId,
		ticketAreaId: items[0].ticketAreaId,
		ticketSaleId: sale.id,
		items,
	};
	const headers = { "idempotency-key": uuidv4() };

	let tries = 0;

	while (tries < 3) {
		const data = group("place order", () => {
			return ticketService.orderRoutesPlaceOrder(payload, headers);
		});

		if (data.response.status >= 500) {
			// retry
			tries++;
			sleep(randomIntBetween(3, 10));
		} else if (data.response.status === 400) {
			// bad request
			// return early
			logger.error(`Bad response for payload ${JSON.stringify(payload)}`);
			return {
				state: "Bad Request",
				order: null,
			};
		} else if (data.response.status === 409) {
			return {
				state: "Retry", // retry due to seat taken
				order: null,
			};
		} else if (data.response.status === 200) {
			return {
				state: "Ok",
				order: data.data.data,
			};
		}
	}

	return {
		state: "Retry", // tries exceed limit. assume just retry later but with different seat request
		order: null,
	};
};

export const payOrder = (
	paymentService: PaymentServiceClient,
	state: ProfileState,
	order: Order,
) => {
	//
};
