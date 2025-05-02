export interface UserProfile {
	dayPreference: "Specific" | "Any Day";
	seatingTier:
		| "Seated-Low"
		| "Seated-Mid"
		| "Seated-High"
		| "Area-Mid"
		| "Area-High";
	quantity: "Solo" | "Couple" | "Group";
	persistence: "Low" | "Medium" | "High";
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
