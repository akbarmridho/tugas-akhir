import { createCluster } from "redis";
import { env } from "./env.js";

export type RedisCluster = ReturnType<typeof createCluster>;

export const redis = createCluster({
	rootNodes: env.REDIS_HOSTS.map((host) => {
		return {
			url: `redis://${host}`,
		};
	}),
});
