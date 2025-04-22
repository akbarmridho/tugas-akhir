import { Queue } from "bullmq";
import type { Cluster } from "ioredis";

export type WebhookQueue = Queue<{ id: string }, void, "webhook">;

export const createWebhookQueue = (redis: Cluster): WebhookQueue => {
	return new Queue<{ id: string }, void, "webhook">("{webhook}", {
		connection: redis,
		defaultJobOptions: {
			backoff: {
				type: "exponential",
				delay: 5000,
			},
			attempts: 10,
		},
	});
};
