import {
	randomIntBetween,
	randomItem,
} from "https://jslib.k6.io/k6-utils/1.2.0/index.js";
import { Event, TicketSale } from "../client/ticket/ticketBackendService";
import { shuffle } from "./random";
import { Customer, generateCustomer } from "./faker";

export type DayPreference = "Specific" | "Any Day";
export type SeatingTier =
	| "Seated-Low"
	| "Seated-Mid"
	| "Seated-High"
	| "Area-Mid"
	| "Area-High";
export type Quantity = "Solo" | "Couple" | "Group";
export type Persistence = "Low" | "Medium" | "High";

export interface UserProfile {
	dayPreference: DayPreference;
	seatingTier: SeatingTier;
	quantity: Quantity;
	persistence: Persistence;
}

const userProfilesData: (UserProfile & { probability: number })[] = [
	{
		dayPreference: "Specific",
		seatingTier: "Seated-Low",
		quantity: "Solo",
		persistence: "Medium",
		probability: 0.01,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Seated-Low",
		quantity: "Solo",
		persistence: "High",
		probability: 0.02,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Seated-Low",
		quantity: "Couple",
		persistence: "Low",
		probability: 0.02,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Seated-Low",
		quantity: "Couple",
		persistence: "Medium",
		probability: 0.03,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Seated-Low",
		quantity: "Couple",
		persistence: "High",
		probability: 0.02,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Seated-Low",
		quantity: "Group",
		persistence: "Low",
		probability: 0.02,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Seated-Low",
		quantity: "Group",
		persistence: "Medium",
		probability: 0.02,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Seated-Mid",
		quantity: "Solo",
		persistence: "Medium",
		probability: 0.01,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Seated-Mid",
		quantity: "Solo",
		persistence: "High",
		probability: 0.02,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Seated-Mid",
		quantity: "Couple",
		persistence: "Low",
		probability: 0.02,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Seated-Mid",
		quantity: "Couple",
		persistence: "Medium",
		probability: 0.02,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Seated-Mid",
		quantity: "Couple",
		persistence: "High",
		probability: 0.02,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Seated-Mid",
		quantity: "Group",
		persistence: "Low",
		probability: 0.03,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Seated-Mid",
		quantity: "Group",
		persistence: "Medium",
		probability: 0.02,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Seated-High",
		quantity: "Solo",
		persistence: "Medium",
		probability: 0.02,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Seated-High",
		quantity: "Solo",
		persistence: "High",
		probability: 0.02,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Seated-High",
		quantity: "Couple",
		persistence: "Medium",
		probability: 0.02,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Seated-High",
		quantity: "Couple",
		persistence: "High",
		probability: 0.02,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Seated-High",
		quantity: "Group",
		persistence: "Medium",
		probability: 0.02,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Area-Mid",
		quantity: "Solo",
		persistence: "Low",
		probability: 0.01,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Area-Mid",
		quantity: "Solo",
		persistence: "Medium",
		probability: 0.02,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Area-Mid",
		quantity: "Couple",
		persistence: "Low",
		probability: 0.01,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Area-Mid",
		quantity: "Couple",
		persistence: "Medium",
		probability: 0.02,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Area-Mid",
		quantity: "Couple",
		persistence: "High",
		probability: 0.01,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Area-Mid",
		quantity: "Group",
		persistence: "Medium",
		probability: 0.02,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Area-High",
		quantity: "Solo",
		persistence: "Medium",
		probability: 0.02,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Area-High",
		quantity: "Solo",
		persistence: "High",
		probability: 0.01,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Area-High",
		quantity: "Couple",
		persistence: "Medium",
		probability: 0.01,
	},
	{
		dayPreference: "Specific",
		seatingTier: "Area-High",
		quantity: "Couple",
		persistence: "High",
		probability: 0.01,
	},
	{
		dayPreference: "Any Day",
		seatingTier: "Seated-Low",
		quantity: "Solo",
		persistence: "Medium",
		probability: 0.02,
	},
	{
		dayPreference: "Any Day",
		seatingTier: "Seated-Low",
		quantity: "Couple",
		persistence: "Low",
		probability: 0.03,
	},
	{
		dayPreference: "Any Day",
		seatingTier: "Seated-Low",
		quantity: "Couple",
		persistence: "Medium",
		probability: 0.03,
	},
	{
		dayPreference: "Any Day",
		seatingTier: "Seated-Low",
		quantity: "Couple",
		persistence: "High",
		probability: 0.03,
	},
	{
		dayPreference: "Any Day",
		seatingTier: "Seated-Low",
		quantity: "Group",
		persistence: "Medium",
		probability: 0.02,
	},
	{
		dayPreference: "Any Day",
		seatingTier: "Seated-Mid",
		quantity: "Solo",
		persistence: "Medium",
		probability: 0.02,
	},
	{
		dayPreference: "Any Day",
		seatingTier: "Seated-Mid",
		quantity: "Couple",
		persistence: "Low",
		probability: 0.03,
	},
	{
		dayPreference: "Any Day",
		seatingTier: "Seated-Mid",
		quantity: "Couple",
		persistence: "Medium",
		probability: 0.01,
	},
	{
		dayPreference: "Any Day",
		seatingTier: "Seated-Mid",
		quantity: "Couple",
		persistence: "High",
		probability: 0.02,
	},
	{
		dayPreference: "Any Day",
		seatingTier: "Seated-Mid",
		quantity: "Group",
		persistence: "Medium",
		probability: 0.04,
	},
	{
		dayPreference: "Any Day",
		seatingTier: "Seated-High",
		quantity: "Solo",
		persistence: "Medium",
		probability: 0.02,
	},
	{
		dayPreference: "Any Day",
		seatingTier: "Seated-High",
		quantity: "Solo",
		persistence: "High",
		probability: 0.02,
	},
	{
		dayPreference: "Any Day",
		seatingTier: "Seated-High",
		quantity: "Group",
		persistence: "Medium",
		probability: 0.03,
	},
	{
		dayPreference: "Any Day",
		seatingTier: "Seated-High",
		quantity: "Couple",
		persistence: "High",
		probability: 0.02,
	},
	{
		dayPreference: "Any Day",
		seatingTier: "Area-Mid",
		quantity: "Solo",
		persistence: "Medium",
		probability: 0.02,
	},
	{
		dayPreference: "Any Day",
		seatingTier: "Area-Mid",
		quantity: "Couple",
		persistence: "Low",
		probability: 0.02,
	},
	{
		dayPreference: "Any Day",
		seatingTier: "Area-Mid",
		quantity: "Couple",
		persistence: "High",
		probability: 0.02,
	},
	{
		dayPreference: "Any Day",
		seatingTier: "Area-Mid",
		quantity: "Group",
		persistence: "Medium",
		probability: 0.02,
	},
	{
		dayPreference: "Any Day",
		seatingTier: "Area-High",
		quantity: "Solo",
		persistence: "High",
		probability: 0.02,
	},
	{
		dayPreference: "Any Day",
		seatingTier: "Area-High",
		quantity: "Couple",
		persistence: "Medium",
		probability: 0.02,
	},
	{
		dayPreference: "Any Day",
		seatingTier: "Area-High",
		quantity: "Couple",
		persistence: "High",
		probability: 0.02,
	},
];

let cumulativeProbability = 0;
const profilesWithCumulative = userProfilesData.map((profile) => {
	cumulativeProbability += profile.probability;
	return { ...profile, cumulative: cumulativeProbability };
});

const totalProbability = cumulativeProbability;

export function getUserProfile(): UserProfile {
	const randomValue = Math.random() * totalProbability;

	for (let i = 0; i < profilesWithCumulative.length; i++) {
		if (randomValue <= profilesWithCumulative[i].cumulative) {
			const { probability, cumulative, ...profileData } =
				profilesWithCumulative[i];
			return profileData;
		}
	}

	// Fallback: Should ideally not be reached if totalProbability > 0
	// Return the last profile if something unexpected happens
	const { probability, cumulative, ...profileData } =
		profilesWithCumulative[profilesWithCumulative.length - 1];
	return profileData;
}

export interface ProfileState {
	saleSelection: TicketSale[];
	ticketCount: number;
	tierOrder: string[];
	customers: Customer[];
	currentBrowseAttempt: number;
	maxBrowseAttempt: number;
	currentOrderAttempt: number;
	maxOrderAttempt: number;
	fallbackType: "Same Day" | "Same Category";
	areaType: "Area" | "Seated";
	saleSkip: Set<number>;
	seatSkip: Set<number>;
	areaSkip: Set<number>;
}

export const getProfileState = (
	profile: UserProfile,
	event: Event,
): ProfileState => {
	const state: ProfileState = {
		saleSelection: [],
		ticketCount: 0,
		tierOrder: [],
		customers: [],
		currentBrowseAttempt: 0,
		maxBrowseAttempt: 0,
		currentOrderAttempt: 0,
		maxOrderAttempt: 0,
		fallbackType: randomItem(["Same Day", "Same Category"]),
		areaType: "Seated",
		saleSkip: new Set(),
		seatSkip: new Set(),
		areaSkip: new Set(),
	};

	// choose the sale selection
	const allSales = event.ticketSales || [];
	shuffle(allSales);

	if (profile.dayPreference === "Any Day") {
		state.saleSelection = allSales;
	} else {
		// for specific day we choose from 1 to at most floor(n/2) day
		const selectedCount = randomIntBetween(1, Math.floor(allSales.length / 2));
		state.saleSelection = allSales.slice(0, selectedCount);
	}

	// choose the ticket count
	if (profile.quantity === "Solo") {
		state.ticketCount = 1;
	} else if (profile.quantity === "Couple") {
		state.ticketCount = 2;
	} else {
		state.ticketCount = randomIntBetween(3, 5);
	}

	for (let i = 0; i < state.ticketCount; i++) {
		state.customers.push(generateCustomer());
	}

	// persistence
	if (profile.persistence === "High") {
		state.maxBrowseAttempt = 27;
		state.maxOrderAttempt = 9;
	} else if (profile.persistence === "Medium") {
		state.maxBrowseAttempt = 18;
		state.maxOrderAttempt = 6;
	} else if (profile.persistence === "Low") {
		state.maxBrowseAttempt = 9;
		state.maxOrderAttempt = 3;
	}

	// tier order
	// first is main second and rest is fallback
	if (profile.seatingTier === "Seated-Low") {
		state.tierOrder = ["Bronze", "Silver"];
	} else if (profile.seatingTier === "Seated-Mid") {
		const secondChoice = ["Gold", "Bronze"];
		shuffle(secondChoice); // shuffle the second choice
		state.tierOrder = ["Silver", ...secondChoice]; // Keep silver as main choice
	} else if (profile.seatingTier === "Seated-High") {
		const choices = ["Platinum", "Gold"];
		shuffle(choices); // shuffle the primary choice
		state.tierOrder = choices;
	} else if (profile.seatingTier === "Area-Mid") {
		state.tierOrder = ["VIP", "Zone A"];
		state.areaType = "Area";
	} else if (profile.seatingTier === "Area-High") {
		const choices = ["Zone A", "Zone B"];
		shuffle(choices); // shuffle the primary choice
		state.tierOrder = choices;
		state.areaType = "Area";
	}

	return state;
};
