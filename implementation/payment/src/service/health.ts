import type { Cluster } from "ioredis";
import type { Logger } from "winston";
import { withTimeout } from "../utils/timeout.js";
import type { ContentfulStatusCode } from "hono/utils/http-status";

interface HealthStatus {
	data: {
		status: string;
		message: string;
	};
	code: ContentfulStatusCode;
}

export class HealthService {
	private redis: Cluster;
	private logger: Logger;

	constructor(redis: Cluster, logger: Logger) {
		this.redis = redis;
		this.logger = logger;
	}

	public async check(): Promise<HealthStatus> {
		try {
			const retrievedValue = await withTimeout(
				this.redis.ping(),
				7000,
				"Redis cluster ping timed out",
			);

			if (retrievedValue === "PONG") {
				return {
					data: {
						status: "healthy",
						message: "Node healthy",
					},
					code: 200,
				};
			}

			return {
				data: {
					status: "unhealthy",
					message: "Redis cluster returned unexpected response",
				},
				code: 500,
			};
		} catch (error) {
			this.logger.error("Redis health check failed", { error });
			return {
				data: {
					status: "unhealthy",
					message: "Redis cluster connection failed",
				},
				code: 503,
			};
		}
	}
}
