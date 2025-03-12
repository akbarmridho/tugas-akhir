import { env } from "./env.js";
import { Cluster, type NatMap } from "ioredis";

const natMap: NatMap = {};

let isLocalhost = false;

const hosts = env.REDIS_HOSTS.map((e, i) => {
	const [host, port] = e.split(":");

	if (host === "localhost") {
		isLocalhost = true;
		natMap[`redis-node-${i}:6379`] = {
			host,
			port: Number(port),
		};
	}

	return {
		host,
		port: Number(port),
	};
});

export const redis = new Cluster(hosts, {
	redisOptions: {
		password: env.REDIS_PASSWORD,
	},
	natMap,
});
