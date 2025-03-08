import { env } from "./env.js";
import { Cluster } from "ioredis";

// todo set no eviction
export const redis = new Cluster(
	env.REDIS_HOSTS.map((e) => {
		const [host, port] = e.split(":");

		return {
			host,
			port: Number(port),
		};
	}),
);
