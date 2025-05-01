import faker from "k6/x/faker";

interface Customer {
	name: string;
	email: string;
}

// generate customer
export const generateCustomer = (): Customer => {
	return {
		email: faker.person.email(),
		name: faker.person.name(),
	};
};
