import type { Env } from "./env.js";
import { Cluster, type NatMap } from "ioredis";

export const newRedisCluster = (config: Env): Cluster => {
	const natMap: NatMap = {};

	if (
		config.REDIS_HOST_MAPS.length > 0 &&
		config.REDIS_HOSTS.length !== config.REDIS_HOST_MAPS.length
	) {
		throw new Error("Length of redis host maps and redis hosts does not match");
	}

	const hosts = config.REDIS_HOSTS.map((e, i) => {
		const [host, port] = e.split(":");

		if (config.REDIS_HOST_MAPS.length > 0) {
			natMap[config.REDIS_HOST_MAPS[i]] = {
				host,
				port: Number(port),
			};
		}

		return {
			host,
			port: Number(port),
		};
	});

	return new Cluster(hosts, {
		redisOptions: {
			password: config.REDIS_PASSWORD,
		},
		natMap,
	});
};
