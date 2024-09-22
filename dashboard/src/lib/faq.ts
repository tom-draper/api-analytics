const faq = [
    {
        question: 'Can the project be self-hosted?',
        answer: 'We are currently working on a self-hosted solution that will allow you to take full control over your logged request data. Check back soon for updates!',
        showing: false,
    },
    {
        question: 'How is this project funded?',
        answer: 'This project is fully funded by myself as a hobby project.',
        showing: false,
    },
    {
        question: 'I have lost my API key, how can I recover it?',
        answer: 'Your API key is the only link between you and your logged requests. We do not store any other forms of authentication such as email addresses. If you have lost your API key, please get in contact with us, and we may be able to recover it for you. If do not access your dashboard or data again, your account and all associated data will be deleted after 3 months, or 6 months if your API continues to log requests.',
        showing: false,
    },
    {
        question: 'Is the service free?',
        answer: 'Account creation is complete free. We may introduce a paid tier in the future if there is demand, but a free tier will always remain available.',
        showing: false,
    },
    {
        question: 'How can I create an account?',
        answer: 'Account creation is complete free, and we do not require any email address. You can generate a free API key <a href="https://www.apianalytics.dev/generate">here</a>.',
        showing: false,
    },
    {
        question: 'How do you comply with GDPR?',
        answer: 'We adhere to GDPR regulations and requirements. Our privacy policy can be found <a href="https://www.apianalytics.dev/privacy-policy">here</a>.',
        showing: false,
    },
    {
        question: 'What are the usage limits?',
        answer: 'In order to keep the service free, we can currently only store up to 1.5 million requests per user. When you hit this limit, old logged requests will be replaced by new requests. If you find this is not enough, we recommend using another service or waiting until our self-hosting solution is ready. We may introduce a paid tier in the future that could support much higher usage limits if there is demand.',
        showing: false,
    },
    {
        question: 'Why are there requests missing from my dashboard?',
        answer: 'In order to keep the service free, we can currently only store up to 1.5 million requests per user. If you have hit this limit, old requests will be replaced with new requests. Occassionally we will have an outage that is reflected in your dashboard. You can track outages <a href="https://www.apianalytics.dev/outages">here</a>.',
        showing: false,
    },
    {
        question: 'Can I contribute to the project?',
        answer: 'The project is open source and contributions, feedback and suggestions are always welcome.',
        showing: false,
    },
    {
        question: 'How long do you store data for?',
        answer: 'Data is only stored for as long as the service may be in use. Any account along with all associated data is scheduled for deletion when either: 3 months have elapsed since the last request was logged by your API, or 6 months have elapsed since you accessed the dashboard or data API, whichever happens first.',
        showing: false,
    },
    {
        question: 'How can I delete my logged request data?',
        answer: 'You can delete your account and all associated data at any time by entering your API key here.',
        showing: false,
    },
    {
        question: 'How can I amend or modify my logged request data?',
        answer: 'We do not support the ability for users to amend or modify their logged request data themselves. If you have a specific request, please get in contact with us.',
        showing: false,
    },
    {
        question: 'How can I request my data?',
        answer: 'You can access your raw logged request data by sending a GET request to our data API.',
        showing: false,
    }
];

export default faq;