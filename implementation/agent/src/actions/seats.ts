import {
	randomIntBetween,
	randomItem,
} from "https://jslib.k6.io/k6-utils/1.2.0/index.js";
import {
	AreaAvailability,
	OrderItemDto,
	TicketBackendServiceClient,
	TicketPackage,
	TicketSale,
	TicketSeat,
} from "../client/ticket/ticketBackendService";
import { ProfileState } from "../utils/profile";
import { group, check } from "k6";

interface AreaSale {
	area: AreaAvailability;
	sale: TicketSale;
}

interface SeatSale {
	sale: TicketSale;
	seats: TicketSeat[];
}

interface OrderItemSale {
	items: OrderItemDto[];
	sale: TicketSale;
}

/**
 *
 * @param ticketService
 * @param ticketSaleId
 * @param state
 * @returns Two dimensional array with first dimension is the order of preference
 */
const getAvailability = (
	ticketService: TicketBackendServiceClient,
	ticketSale: TicketSale,
	state: ProfileState,
): AreaAvailability[][] | null => {
	return group("get availability", () => {
		const response = ticketService.eventRoutesGetAvailability(ticketSale.id, {
			tags: {
				name: "GetAvailability",
			},
		});

		const ok = check(response, {
			"availability: is status 200": (r) => r.response.status === 200,
			"availability: not empty": (r) => r.data.data.length !== 0,
		});

		if (!ok) {
			return null;
		}

		const data = response.data.data;

		if (data.length === 0) {
			return null;
		}

		// filter by category preference

		const packageMap: { [key: number]: TicketPackage | undefined } =
			ticketSale.ticketPackages.reduce(
				(prev, curr) => {
					prev[curr.id] = curr;
					return prev;
				},
				{} as { [key: number]: TicketPackage | undefined },
			);

		const result: AreaAvailability[][] = [];

		for (const category of state.tierOrder) {
			result.push(
				data.filter((e) => {
					const ticketPackage = packageMap[e.ticketPackageId];

					if (!ticketPackage) {
						return false;
					}

					return (
						!state.areaSkip.has(e.ticketAreaId) && // area does not skipped
						e.availableSeats >= state.ticketCount && // available seats more than want to buy
						ticketPackage.ticketCategory.name.includes(category)
					);
				}),
			);
		}

		return result;
	});
};

const getSeats = (
	ticketService: TicketBackendServiceClient,
	state: ProfileState,
	area: AreaAvailability,
): TicketSeat[] | null => {
	return group("get seats", () => {
		const response = ticketService.eventRoutesGetSeats(area.ticketAreaId, {
			tags: {
				name: "GetSeat",
			},
		});

		const ok = check(response, {
			"seats: is status 200": (r) => r.response.status === 200,
			"seats: not empty": (r) => r.data.data.length !== 0,
		});

		if (!ok) {
			return null;
		}

		const data = response.data.data;

		if (data.length === 0) {
			return null;
		}

		// filter status
		// assume consecutive id correlates to consecutive area number
		const availableSeats = data
			.filter((seat) => seat.status === "available")
			.map((e) => {
				const splitted = e.seatNumber.split("-");
				const seatNumber = parseInt(splitted[splitted.length - 1]);
				return {
					...e,
					parsedNumber: seatNumber,
				};
			})
			.sort((a, b) => a.parsedNumber - b.parsedNumber);

		const options = availableSeats
			.reduce(
				(prev, curr) => {
					if (prev.length === 0) {
						prev.push([curr]);
						return prev;
					}

					const last = prev[prev.length - 1];
					const lastSeat = last[last.length - 1];

					// not consecutive
					if (Math.abs(lastSeat.parsedNumber - curr.parsedNumber) > 1) {
						prev.push([curr]);
					} else {
						last.push(curr);
					}

					return prev;
				},
				[] as (TicketSeat & { parsedNumber: number })[][],
			)
			// user want consecutive
			.filter((option) => option.length >= state.ticketCount);

		if (options.length === 0) {
			return null;
		}

		const selectedOption = randomItem(options);

		// options 8
		// want 3
		// idx can be from 0 to 5
		const startIdx = randomIntBetween(
			0,
			selectedOption.length - state.ticketCount,
		);

		const selectedSeats = selectedOption.slice(
			startIdx,
			startIdx + state.ticketCount,
		);

		check(selectedSeats, {
			"selected-seats: selected seat and want equal": (s) =>
				s.length === state.ticketCount,
		});

		return selectedSeats;
	});
};

const selectTicketSale = (state: ProfileState): TicketSale | null => {
	const options = state.saleSelection.filter((s) => !state.areaSkip.has(s.id));
	if (options.length === 0) {
		return null;
	}

	return randomItem(options);
};

const getAreaPreferSameDay = (
	ticketService: TicketBackendServiceClient,
	state: ProfileState,
): AreaSale | null => {
	// check same day first
	const ticketSale = selectTicketSale(state);

	if (!ticketSale) {
		return null;
	}

	const availabilities = getAvailability(ticketService, ticketSale, state);

	if (availabilities === null || availabilities.length === 0) {
		return null;
	}

	for (const areas of availabilities) {
		if (areas.length === 0) {
			continue;
		}

		return {
			area: randomItem(areas),
			sale: ticketSale,
		};
	}

	return null;
};

const getAreaPreferSameCategory = (
	ticketService: TicketBackendServiceClient,
	state: ProfileState,
): AreaSale | null => {
	const saleAreaAvailability: (AreaAvailability[][] | null)[] = [];

	for (let catIdx = 0; catIdx < state.tierOrder.length; catIdx++) {
		for (let saleIdx = 0; saleIdx < state.saleSelection.length; saleIdx++) {
			if (catIdx === 0) {
				// fetch the availability
				saleAreaAvailability.push(
					getAvailability(ticketService, state.saleSelection[saleIdx], state),
				);
			}

			// area availability
			const areaAvailability = saleAreaAvailability[saleIdx];

			if (areaAvailability === null || areaAvailability.length === 0) {
				continue;
			}

			const areas = areaAvailability[catIdx];

			if (areas.length === 0) {
				continue;
			}

			return {
				area: randomItem(areas),
				sale: state.saleSelection[saleIdx],
			};
		}
	}

	return null;
};

const getSeatCaseArea = (
	ticketService: TicketBackendServiceClient,
	state: ProfileState,
): AreaSale | null => {
	while (state.currentBrowseAttempt < state.maxBrowseAttempt) {
		let area: AreaSale | null = null;

		if (state.fallbackType === "Same Category") {
			area = getAreaPreferSameCategory(ticketService, state);
		} else {
			area = getAreaPreferSameDay(ticketService, state);
		}

		if (!area) {
			state.currentBrowseAttempt++;
			continue;
		}

		return area;
	}

	return null;
};

const getSeatCaseNumbered = (
	ticketService: TicketBackendServiceClient,
	state: ProfileState,
): SeatSale | null => {
	while (state.currentBrowseAttempt < state.maxBrowseAttempt) {
		const areaSale = getSeatCaseArea(ticketService, state);

		if (!areaSale) {
			state.currentBrowseAttempt++;
			continue;
		}

		// try at most 3 times
		for (let i = 0; i < 3; i++) {
			const seats = getSeats(ticketService, state, areaSale.area);

			if (!seats) {
				continue;
			}

			return {
				sale: areaSale.sale,
				seats: seats,
			};
		}

		return null;
	}

	return null;
};

export const getSeat = (
	ticketService: TicketBackendServiceClient,
	state: ProfileState,
): OrderItemSale | null => {
	if (state.areaType === "Area") {
		// area
		const areaSale = getSeatCaseArea(ticketService, state);

		if (!areaSale) {
			return null;
		}

		const items: OrderItemDto[] = [];

		for (let i = 0; i < state.ticketCount; i++) {
			const customer = state.customers[i];
			items.push({
				customerEmail: customer.email,
				customerName: customer.name,
				ticketAreaId: areaSale.area.ticketAreaId,
			});
		}

		return {
			items,
			sale: areaSale.sale,
		};
	}

	// seated
	const seatSale = getSeatCaseNumbered(ticketService, state);

	if (!seatSale) {
		return null;
	}

	const items: OrderItemDto[] = [];

	for (let i = 0; i < state.ticketCount; i++) {
		const customer = state.customers[i];
		const seat = seatSale.seats[i];
		items.push({
			customerEmail: customer.email,
			customerName: customer.name,
			ticketAreaId: seat.ticketAreaId,
			ticketSeatId: seat.id,
		});
	}

	return {
		items,
		sale: seatSale.sale,
	};
};
