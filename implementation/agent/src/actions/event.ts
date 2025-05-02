import {
	Event,
	TicketBackendServiceClient,
} from "../client/ticket/ticketBackendService";
import { group, check } from "k6";

export const getEvent = (
	ticketService: TicketBackendServiceClient,
): Event | null => {
	return group("test", () => {
		const eventsResponse = ticketService.eventRoutesGetEvents();

		const eventsCheck = check(eventsResponse, {
			"is status 200": (r) => r.response.status === 200,
			"length is 1": (r) => r.data.data.length === 1,
		});

		if (!eventsCheck) {
			return null;
		}

		const id = eventsResponse.data.data[0].id;

		const eventResponse = ticketService.eventRoutesGetEvent(id);

		const eventCheck = check(eventResponse, {
			"is status 200": (r) => r.response.status === 200,
		});

		if (!eventCheck) {
			return null;
		}

		return eventResponse.data.data;
	});
};
