import { check, group, sleep } from "k6";
import {
	Order,
	OrderItemDto,
	TicketBackendServiceClient,
	TicketSale,
} from "../client/ticket/ticketBackendService";
import { ProfileState } from "../utils/profile";
import { uuidv4 } from "https://jslib.k6.io/k6-utils/1.4.0/index.js";
import { randomIntBetween } from "https://jslib.k6.io/k6-utils/1.2.0/index.js";
import { logger } from "../utils/logger";
import {
	Invoice,
	PaymentServiceClient,
} from "../client/payment/paymentService";

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
			return ticketService.orderRoutesPlaceOrder(payload, headers, {
				tags: {
					name: "PlaceOrder",
				},
			});
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
	order: Order,
	paymentSuccess: boolean,
): Invoice | null => {
	return group("pay invoice", () => {
		const response = paymentService.postInvoicesIdPayment(
			`${order.invoice!.externalId}`,
			{
				mode: paymentSuccess ? "success" : "failed",
			},
			{
				tags: {
					name: "PayInvoice",
				},
			},
		);

		const ok = check(response, {
			"pay-invoice: is status 200": (r) => r.response.status === 200,
		});

		if (!ok) {
			logger.error(
				`Pay invoice got error ${response.response.status} ${JSON.stringify(response.response.json())}`,
			);

			return null;
		}

		return response.data;
	});
};
