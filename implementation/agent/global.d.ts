declare module "https://jslib.k6.io/k6-utils/1.4.0/index.js" {
	export const uuidv4: () => string;
}

declare module "https://jslib.k6.io/k6-utils/1.2.0/index.js" {
	export const randomIntBetween: (a: number, b: number) => number;
	export const randomItem: <T>(arr: T[]) => T;
}
