import { Worker } from "bullmq";
import { env } from "../../infrastructure/env.js";
import { InvoiceSchema } from "../../entity/invoice.js";
import got from "got";
import { readFileSync } from "node:fs";
import { computeHMACSHA256 } from "../../utils/hmac.js";
import { newRedisCluster } from "../../infrastructure/redis.js";
import { createLogger } from "../../utils/logger.js";
import { InvoiceService } from "../../service/invoice.js";
import { IdGenerator } from "../../utils/id-generator.js";

(async function main() {
	const idGenerator = new IdGenerator();
	const logger = createLogger(env);
	const redis = newRedisCluster(env);
	const invoiceService = new InvoiceService(redis, logger, idGenerator);

	// Create the worker to process webhook jobs
	const worker = new Worker<{ id: string }, void, "webhook">(
		"{webhook}",
		async (job) => {
			const { id } = job.data;

			const l = logger.child({ id });
			l.info("handling job");

			let invoice = await invoiceService.getInvoice(id);

			if (invoice === null) {
				throw new Error("Invoice Not Found");
			}

			if (invoice.status === "pending") {
				// status still pending, which mean it's expired
				const newInvoice = await invoiceService.expireInvoice(id);

				if (newInvoice === null) {
					throw new Error("Expire invoice result must not be null");
				}

				invoice = newInvoice;
			}

			const payload = JSON.stringify(invoice);
			const hash = computeHMACSHA256(env.WEBHOOK_SECRET, payload);

			try {
				const response = await got.post(env.WEBHOOK_URL, {
					body: payload,
					headers: {
						"content-type": "application/json",
						"x-webhook-verify": hash,
					},
					https: {
						key: readFileSync(env.KEY_PATH),
						certificate: readFileSync(env.CERT_PATH),
						rejectUnauthorized: false,
					},
					http2: true,
					throwHttpErrors: false,
				});

				if (response.ok) {
					l.info(`Webhook success ${id}`);
				} else if (response.statusCode === 404) {
					l.warn("Received not found error. ignoring ...");
				} else if (response.statusCode >= 400) {
					const statusCode = response.statusCode;

					let errorBody: string;
					try {
						const body = JSON.parse(response.body);

						// biome-ignore lint/suspicious/noExplicitAny: <explanation>
						if (Object.hasOwn(body as any, "message")) {
							// biome-ignore lint/suspicious/noExplicitAny: <explanation>
							errorBody = (body as any).message;
						} else {
							errorBody = JSON.stringify(response.body);
						}
					} catch {
						errorBody = "Could not read error response body";
					}

					// For other status codes, throw a generic error that will trigger retry
					throw new Error(
						`HTTP Error: ${statusCode} ${response.statusMessage} with body ${errorBody}`,
					);
				}
			} catch (error) {
				// Log all errors (both network errors and HTTP errors)
				l.error(`Webhook error: ${(error as Error).message}`, {
					error,
				});

				// Re-throw to trigger retry
				throw error;
			}
		},
		{
			connection: redis,
			concurrency: 500, // Process 500 webhooks simultaneously
		},
	);

	// Graceful shutdown
	async function shutdown() {
		await worker.close();
		process.exit(0);
	}

	process.on("SIGTERM", shutdown);
	process.on("SIGINT", shutdown);
})().catch((e) => {
	console.error("Error", e);
});
