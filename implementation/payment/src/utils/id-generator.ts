import { init } from "@paralleldrive/cuid2";
import { randomBytes } from "node:crypto";

export class IdGenerator {
	private instanceId: string;
	private generator: () => string;

	constructor() {
		this.instanceId = randomBytes(20).toString("hex");
		this.generator = init({
			length: 15,
			fingerprint: this.instanceId,
		});
	}

	public generate(): string {
		return this.generator();
	}

	public getInstanceId(): string {
		return this.instanceId;
	}
}
