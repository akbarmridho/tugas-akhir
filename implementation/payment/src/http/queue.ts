import { Queue } from "bullmq";
import { redis } from "../common/redis.js";

export const queue = new Queue<{ id: string }, void, "webhook">("{webhook}", {
	connection: redis,
	defaultJobOptions: {
		backoff: {
			type: "exponential",
			delay: 5000,
		},
		attempts: 10,
	},
});
