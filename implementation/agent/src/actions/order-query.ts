import { check, group } from "k6";
import {
	IssuedTicket,
	Order,
	TicketBackendServiceClient,
} from "../client/ticket/ticketBackendService";
import { logger } from "../utils/logger";

export const getOrder = (
	ticketService: TicketBackendServiceClient,
	id: number,
): Order | null => {
	return group("get order", () => {
		const orderResponse = ticketService.orderRoutesGetOrder(id, {
			tags: {
				name: "GetOrder",
			},
		});

		const orderCheck = check(orderResponse, {
			"is status 200": (r) => r.response.status === 200,
		});

		if (!orderCheck) {
			logger.error(
				`get order got status ${orderResponse.response.status} ${orderResponse.response.json()}`,
			);
			return null;
		}

		return orderResponse.data.data;
	});
};

export const getIssuedTickets = (
	ticketService: TicketBackendServiceClient,
	id: number,
): IssuedTicket[] | null => {
	return group("get issued tickets", () => {
		const orderResponse = ticketService.orderRoutesGetIssuedTickets(id, {
			tags: {
				name: "GetIssuedTickets",
			},
		});

		const orderCheck = check(orderResponse, {
			"is status 200": (r) => r.response.status === 200,
		});

		if (!orderCheck) {
			logger.error(
				`get issued tickets got status ${orderResponse.response.status} ${orderResponse.response.json()}`,
			);
			return null;
		}

		return orderResponse.data.data;
	});
};
