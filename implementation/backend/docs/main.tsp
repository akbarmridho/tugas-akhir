import "@typespec/http";
import "@typespec/rest";
import "@typespec/openapi3";

using TypeSpec.Http;
using TypeSpec.Rest;

@service(#{ title: "Ticket Backend Service" })
@server(
    "{protocol}://{hostname}",
    "Server host",
    {
        protocol: "http" | "https" = "https",
        hostname: string = "localhost:3000",
    }
)
namespace Main;

/**
 * Responses
 */
alias PayloadResponse<Type> = {
    @statusCode statusCode: 200;
    @header contentType: "application/json";
    @body payload: {
        data: Type;
        message?: string;
    };
};

model FieldError {
    field: string;
    tag: string;
    message: string;
}

@error
model ErrorResponseModel {
    message: string;
    errors?: FieldError[];
}

model InputErrorResponse {
    @statusCode statusCode: 400;
    @header contentType: "application/json";
    @body error: ErrorResponseModel;
}

model ResourceNotFoundResponse {
    @statusCode statusCode: 404;
    @header contentType: "application/json";
    @body error: ErrorResponseModel;
}

model ResourceConflictResponse {
    @statusCode statusCode: 409;
    @header contentType: "application/json";
    @body data: {
        message: string;
    };
}

/**
 * Events
 */
enum SeatStatus {
    Available: "available",
    OnHold: "on-hold",
    Sold: "sold",
}

enum AreaType {
    NumberedSeating: "numbered-seating",
    FreeStanding: "free-standing",
}

model AreaAvailability {
    ticketPackageId: int64;
    ticketAreaId: int64;
    totalSeats: int32;
    availableSeats: int32;
}

model TicketArea {
    id: int64;
    type: AreaType;
    ticketPackageId: int64;
    createdAt: utcDateTime;
    updatedAt: utcDateTime;
    ticketSeats?: TicketSeat[];
}

model TicketSeat {
    id: int64;
    seatNumber: string;
    status: SeatStatus;
    ticketAreaId: int64;
    createdAt: utcDateTime;
    updatedAt: utcDateTime;
    ticketArea?: TicketArea;
}

model Event {
    id: int64;
    name: string;
    location: string;
    description: string;
    createdAt: utcDateTime;
    updatedAt: utcDateTime;
    ticketSales?: TicketSale[];
}

model TicketCategory {
    id: int64;
    name: string;
    createdAt: utcDateTime;
    updatedAt: utcDateTime;
}

model TicketSale {
    id: int64;
    name: string;
    saleBeginAt: utcDateTime;
    saleEndAt: utcDateTime;
    eventId: int64;
    createdAt: utcDateTime;
    updatedAt: utcDateTime;

    // Relations
    ticketPackages: TicketPackage[];
}

model TicketPackage {
    id: int64;
    price: int32;
    ticketCategoryId: int64;
    ticketSaleId: int64;
    createdAt: utcDateTime;
    updatedAt: utcDateTime;

    // Relations
    ticketCategory: TicketCategory;

    ticketAreas: TicketArea[];
}

model GetEventDto {
    @path
    eventId: int64;
}

model GetAvailabilityDto {
    @path
    ticketSaleId: int64;
}

model GetSeatsDto {
    @path
    ticketAreaId: int64;
}

/**
 * Orders
 */
enum OrderStatus {
    WaitingForPayment: "waiting-for-payment",
    Failed: "failed",
    Success: "success",
}

enum InvoiceStatus {
    Pending: "pending",
    Expired: "expired",
    Failed: "failed",
    Success: "success",
}

model Invoice {
    id: int64;
    status: InvoiceStatus;
    amount: int32;
    externalId: string;
    orderId: int64;
    createdAt: utcDateTime;
    updatedAt: utcDateTime;
}

model Order {
    id: int64;
    status: OrderStatus;
    failReason: string | null;
    eventId: int64;
    ticketSaleId: int64;
    ticketAreaId: int64;
    externalUserId: string;
    createdAt: utcDateTime;
    updatedAt: utcDateTime;

    // relations
    items: OrderItem[];

    invoice: Invoice | null;
    event?: Event;
    ticketSale?: TicketSale;
}

model OrderItem {
    id: int64;
    customerName: string;
    customerEmail: string;
    price: int64;
    orderId: int64;
    ticketCategoryId: int64;
    ticketSeatId: int64;
    createdAt: utcDateTime;
    updatedAt: utcDateTime;

    // relations
    ticketSeat?: TicketSeat;

    ticketCategory?: TicketCategory;
}

model OrderItemDto {
    customerName: string;
    customerEmail: string;
    ticketSeatId?: int64;
    ticketAreaId: int64;
}

model PlaceOrderDto {
    eventId: int64;
    ticketSaleId: int64;
    ticketAreaId: int64;

    @minItems(1)
    @maxItems(5)
    items: OrderItemDto[];
}

model IssuedTicket {
    id: int64;
    serialNumber: string;
    holderName: string;
    name: string;
    description: string;
    seatId: int64;
    orderId: int64;
    orderItemId: int64;
    createdAt: utcDateTime;
    updatedAt: utcDateTime;

    // Relations
    ticketSeat: TicketSeat;
}

model GetOrderDto {
    @path
    id: int64;
}

model GetIssuedTicketDto {
    @path
    id: int64;
}

/**
 * Main Routes
 */
@route("/events")
@useAuth(BearerAuth)
@tag("events")
namespace EventRoutes {
    @route("/availability/{ticketSaleId}")
    @get
    op getAvailability(
        ...GetAvailabilityDto,
    ): PayloadResponse<AreaAvailability[]> | InputErrorResponse | ResourceNotFoundResponse;

    @route("/seats/{ticketAreaId}")
    @get
    op getSeats(
        ...GetSeatsDto,
    ): PayloadResponse<TicketSeat[]> | InputErrorResponse | ResourceNotFoundResponse;

    @route("/{eventId}")
    @get
    op getEvent(
        ...GetEventDto,
    ): PayloadResponse<Event> | InputErrorResponse | ResourceNotFoundResponse;

    @route("/")
    @get
    op getEvents(): PayloadResponse<Event[]>;
}

@route("/orders")
@useAuth(BearerAuth)
@tag("orders")
namespace OrderRoutes {
    @route("/")
    @post
    op placeOrder(
        @header idempotencyKey: string,
        @body body: PlaceOrderDto,
    ): PayloadResponse<Order> | ResourceConflictResponse | InputErrorResponse;

    @route("/{id}")
    @get
    op getOrder(
        ...GetOrderDto,
    ): PayloadResponse<Order> | ResourceNotFoundResponse | InputErrorResponse;

    @route("/{id}/tickets")
    @get
    op getIssuedTickets(
        ...GetIssuedTicketDto,
    ): PayloadResponse<IssuedTicket[]> | ResourceNotFoundResponse | InputErrorResponse;
}
