import http from 'k6/http';
import {check, fail} from "k6";
import {Counter, Trend} from "k6/metrics";


export const options = {
    scenarios: {
        my_scenario1: {
            executor: 'constant-arrival-rate',
            duration: '4m',
            preAllocatedVUs: 600,
            rate: 600,
            timeUnit: '1s',
        },
    },
};

const BASE_URL = 'http://ip-172-31-23-213.ap-southeast-1.compute.internal'

const TICKET_BASE_URL = `${BASE_URL}/api/ticket/ex`
const ORDER_BASE_URL = `${BASE_URL}/api/order/ex`
const PAYMENT_BASE_URL = `${BASE_URL}/api/payment/ex`
const TICKET_GROUP_ID = '01HJKCYC8QCNHNTJHE6RWYTDJ7'

const getTicketCategoriesSuccessCounter = new Counter('custom_get_ticket_categories_success');
const getTicketsSuccessCounter = new Counter('custom_get_tickets_success');
const createOrderSuccessCounter = new Counter('custom_create_order_success');
const createPaymentSuccessCounter = new Counter('custom_create_payment_success');

const getTicketCategoriesFailedCounter = new Counter('custom_get_ticket_categories_failed');
const getTicketsFailedCounter = new Counter('custom_get_tickets_failed');
const createOrderFailedCounter = new Counter('custom_create_order_failed');
const createPaymentFailedCounter = new Counter('custom_create_payment_failed');

const getTicketCategoriesWarnCounter = new Counter('custom_get_ticket_categories_warn');
const getTicketsWarnCounter = new Counter('custom_get_tickets_warn');
const createOrderWarnCounter = new Counter('custom_create_order_warn');
const createPaymentWarnCounter = new Counter('custom_create_payment_warn');

const getTicketCategoriesTimeoutCounter = new Counter('custom_get_ticket_categories_timeout');
const getTicketsTimeoutCounter = new Counter('custom_get_tickets_timeout');
const createOrderTimeoutCounter = new Counter('custom_create_order_timeout');
const createPaymentTimeoutCounter = new Counter('custom_create_payment_timeout');

const getTicketCategoriesWaiting = new Trend('custom_get_ticket_categories_waiting');
const getTicketsWaiting = new Trend('custom_get_tickets_waiting');
const createOrderWaiting = new Trend('custom_create_order_waiting');
const createPaymentWaiting = new Trend('custom_create_payment_waiting');

const params = {
    headers: {
        'Content-Type': 'application/json',
        'X-User-Id': '01HK6PWF0BA0SGWZSBAHXEPNVH',
        'X-User-Email': 'arya@test.com'
    },
};

export default function () {
    const categories = getTicketCategories();
    const categoryRand = Math.floor(Math.random() * categories.length);
    const category = categories[categoryRand].category;

    const tickets = getTickets(category);
    const ticketRand = Math.floor(Math.random() * tickets.length);
    const ticketId = tickets[ticketRand].id;

    const orderId = createOrder(ticketId);

    const doPayAction = __VU % 2 === 0;
    if (doPayAction) {
        createPayment(orderId);
    }
}

function getTicketCategories() {
    const getTicketCategories = http.get(`${TICKET_BASE_URL}/v1/ticket-groups/${TICKET_GROUP_ID}/categories`, params);
    getTicketCategoriesWaiting.add(getTicketCategories.timings.waiting);

    let isError = check(getTicketCategories, {
        'is 500': (r) => r.status >= 500,
    });

    if (isError) {
        getTicketCategoriesFailedCounter.add(1);
        fail('Internal error');
    }

    const checkTimeoutGetTicketCategories = check(getTicketCategories, {
        'is status 408': (r) => r.status === 408,
    });

    if (checkTimeoutGetTicketCategories) {
        getTicketCategoriesTimeoutCounter.add(1);
        fail('Timeout get ticket categories');
    }

    const checkGetTicketCategories = check(getTicketCategories, {
        'is status 200': (r) => r.status === 200,
    });

    if (!checkGetTicketCategories) {
        getTicketCategoriesWarnCounter.add(1);
        fail('Warn check ticket categories');
    }

    getTicketCategoriesSuccessCounter.add(1);
    return getTicketCategories.json('data');
}

function getTickets(category) {
    const getTickets = http.get(`${TICKET_BASE_URL}/v1/ticket-groups/${TICKET_GROUP_ID}?category=${category}`, params);
    getTicketsWaiting.add(getTickets.timings.waiting);
    let isError = check(getTickets, {
        'is 500': (r) => r.status >= 500,
    });

    if (isError) {
        console.log(category);
        getTicketsFailedCounter.add(1);
        fail('Internal error');
    }

    const checkTimeoutGetTickets = check(getTickets, {
        'is status 408': (r) => r.status === 408,
    });

    if (checkTimeoutGetTickets) {
        getTicketsTimeoutCounter.add(1);
        fail('Timeout get tickets');
    }

    const checkGetTickets = check(getTickets, {
        'is status 200': (r) => r.status === 200,
    });

    if (!checkGetTickets) {
        getTicketsWarnCounter.add(1);
        fail('Warn get tickets');
    }

    getTicketsSuccessCounter.add(1);
    return getTickets.json('data');
}

function createOrder(ticketId) {
    const createOrderPayload = JSON.stringify({"ticket_id": ticketId})

    const createOrderResponse = http.post(`${ORDER_BASE_URL}/v1/orders`, createOrderPayload, params);
    createOrderWaiting.add(createOrderResponse.timings.waiting);
    let isError = check(createOrderResponse, {
        'is 500': (r) => r.status >= 500,
    });

    if (isError) {
        console.log(ticketId);
        createOrderFailedCounter.add(1);
        fail('Internal error');
    }

    const checkTimeoutCreateOrder = check(createOrderResponse, {
        'is status 408': (r) => r.status === 408,
    });

    if (checkTimeoutCreateOrder) {
        createOrderTimeoutCounter.add(1);
        fail('Timeout create order');
    }

    const checkCreateOrder = check(createOrderResponse, {
        'is status 200': (r) => r.status === 200,
    });

    if (!checkCreateOrder) {
        createOrderWarnCounter.add(1);
        fail('Warn create order');
    }
    createOrderSuccessCounter.add(1);

    return createOrderResponse.json('data.id');
}

function createPayment(orderId) {
    const createPaymentPayload = JSON.stringify({
        "order_id": orderId,
        "card_number": "4242424242424242",
        "exp_month": "12",
        "exp_year": "2020",
        "cvv": "123"
    })

    const createPaymentResponse = http.post(`${PAYMENT_BASE_URL}/v1/payments`, createPaymentPayload, params);
    createPaymentWaiting.add(createPaymentResponse.timings.waiting);
    let isError = check(createPaymentResponse, {
        'is 500': (r) => r.status >= 500,
    });

    if (isError) {
        console.log(orderId);
        createPaymentFailedCounter.add(1);
        fail('Internal error');
    }

    const checkTimeoutCreatePayment = check(createPaymentResponse, {
        'is status 408': (r) => r.status === 408,
    });

    if (checkTimeoutCreatePayment) {
        createPaymentTimeoutCounter.add(1);
        fail('Timeout create payment');
    }

    const checkCreatePayment = check(createPaymentResponse, {
        'is status 200': (r) => r.status === 200,
    });

    if (!checkCreatePayment) {
        createPaymentWarnCounter.add(1);
        fail('Warn create payment');
    }
    createPaymentSuccessCounter.add(1);
}