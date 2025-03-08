import { Queue } from "bullmq";
import { redis } from "../common/redis.js";

export const queue = new Queue("{webhook}", {
	connection: redis,
});
