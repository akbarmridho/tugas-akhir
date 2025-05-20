import {
	Event,
	TicketBackendServiceClient,
} from "../client/ticket/ticketBackendService";
import { group, check } from "k6";

export const getEvent = (
	ticketService: TicketBackendServiceClient,
): Event | null => {
	return group("get event", () => {
		const eventsResponse = ticketService.eventRoutesGetEvents({
			tags: {
				name: "GetEvents",
			},
		});

		const eventsCheck = check(eventsResponse, {
			"events: is status 200": (r) => r.response.status === 200,
		});

		if (!eventsCheck) {
			return null;
		}

		const id = eventsResponse.data.data[0].id;

		const eventResponse = ticketService.eventRoutesGetEvent(id, {
			tags: {
				name: "GetEvent",
			},
		});

		const eventCheck = check(eventResponse, {
			"event: is status 200": (r) => r.response.status === 200,
		});

		if (!eventCheck) {
			return null;
		}

		return eventResponse.data.data;
	});
};
