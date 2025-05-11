import type { Cluster } from "ioredis";
import {
	type CreateInvoiceType,
	InvoiceSchema,
	type InvoiceType,
	type PayInvoiceType,
} from "../entity/invoice.js";
import type { Logger } from "winston";
import { HTTPException } from "hono/http-exception";
import type { IdGenerator } from "../utils/id-generator.js";
import dayjs from "dayjs";

export class InvoiceService {
	private redis: Cluster;
	private logger: Logger;
	private idGenerator: IdGenerator;

	constructor(redis: Cluster, logger: Logger, idGenerator: IdGenerator) {
		this.redis = redis;
		this.logger = logger;
		this.idGenerator = idGenerator;
	}

	public async getInvoice(id: string): Promise<InvoiceType | null> {
		const data = await this.redis.get(`invoices:${id}`);

		if (data === null) {
			return null;
		}

		const invoice = await InvoiceSchema.safeParseAsync(JSON.parse(data));

		if (!invoice.success) {
			throw new HTTPException(500, {
				message: "Cannot parse invoice data from redis",
			});
		}

		return invoice.data;
	}

	public async createInvoice(payload: CreateInvoiceType): Promise<InvoiceType> {
		const id = this.idGenerator.generate();
		const now = dayjs();
		const expireDate = now.add(15, "minute");

		const data: InvoiceType = {
			id,
			amount: payload.amount,
			description: payload.description,
			externalId: payload.externalId,
			createdAt: now.toDate(),
			expiredAt: expireDate.toDate(),
			paidAt: null,
			paidAmount: null,
			status: "pending",
		};

		await this.redis.setex(
			`invoices:${id}`,
			5 * 60 * 60 * 1000,
			JSON.stringify(data),
		);

		return data;
	}

	public async expireInvoice(
		id: string,
	): Promise<InvoiceType | null> {
		const invoice = await this.getInvoice(id);

		if (!invoice) {
			return null;
		}

		if (
			invoice.status !== "pending"
		) {
			throw new HTTPException(400, {
				message: "Invoice must not be not be pending",
			});
		}

		invoice.status = "expired"
	
		await this.redis.setex(
			`invoices:${id}`,
			5 * 60 * 60 * 1000,
			JSON.stringify(invoice),
		);

		return invoice;
	}

	public async payInvoice(
		id: string,
		payload: PayInvoiceType,
	): Promise<InvoiceType | null> {
		const invoice = await this.getInvoice(id);

		if (!invoice) {
			return null;
		}

		if (
			invoice.status !== "pending" ||
			dayjs(invoice.expiredAt).isBefore(dayjs())
		) {
			throw new HTTPException(400, {
				message: "Invoice must not be expired and must not be pending",
			});
		}

		if (payload.mode === "success") {
			invoice.paidAmount = invoice.amount;
			invoice.paidAt = new Date();
			invoice.status = "paid";
		} else {
			invoice.status = "failed";
		}

		await this.redis.setex(
			`invoices:${id}`,
			5 * 60 * 60 * 1000,
			JSON.stringify(invoice),
		);

		return invoice;
	}
}
