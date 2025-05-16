const faq = [
    {
        question: 'Can the project be self-hosted?',
        answer: 'The project can be easily self-hosted following the guide <a href="https://github.com/tom-draper/api-analytics/tree/main/server/self-hosting">here</a>. Self-hosting is still undergoing testing, development and further improvements to make it as easy as possible to deploy. It is currently recommended that you avoid self-hosting for production use.',
        showing: false,
    },
    {
        question: 'How is this project funded?',
        answer: 'This project is funded by myself as a hobby project. Donations are always welcome and greatly appreciated, and will go towards server costs and further development.<br><br><a href="https://www.buymeacoffee.com/tomdraper">Donate</a>',
        showing: false,
    },
    {
        question: 'I have lost my API key, how can I recover it?',
        answer: 'Your API key is the only link between you and your logged requests. We do not store any other forms of authentication such as email addresses. If you have lost your API key, please get in contact with us, and we may be able to recover it for you. If do not access your dashboard or data again, all data associated with the account will be deleted after 3 months.',
        showing: false,
    },
    {
        question: 'Is the service free?',
        answer: 'The service is currently entirely free. In future, we may transition to a fremium-based service with the introduction of a paid tier for higher usage, but a free tier will always remain available.',
        showing: false,
    },
    {
        question: 'How can I create an account?',
        answer: 'Account creation is complete free and takes a second to do. Accounts are controlled entirely through a unique randomly-generated API key. Get your API key <a href="/generate">here</a>.',
        showing: false,
    },
    {
        question: 'Do you comply with GDPR?',
        answer: 'We adhere to GDPR regulations and requirements. All data is stored securely on London-based servers. Data is retained only for the duration necessary to deliver this service and is used solely for that purpose. You may delete your stored data at any time. It is never shared with or sold to any other third parties. Our privacy policy can be found <a href="/privacy-policy">here</a>.',
        showing: false,
    },
    {
        question: 'What are the usage limits?',
        answer: 'In order to keep the service free, we can currently only store up to 1.5 million requests per user. When you hit this limit, old logged requests will be replaced by new requests. If you find this is not enough, we recommend trying another service following the <a href="https://github.com/tom-draper/api-analytics/tree/main/server/self-hosting">self-hosting guide</a>. We may introduce a paid tier in the future that could support much higher usage limits if there is demand.',
        showing: false,
    },
    {
        question: 'Why are there requests missing from my dashboard?',
        answer: 'In order to keep the service free, we can currently only store up to 1.5 million requests per user. If you have hit this limit, old requests will be replaced with new requests.<br><br>Occassionally we will have an outage that is reflected in your dashboard. You can track outages <a href="/outages">here</a>.',
        showing: false,
    },
    {
        question: 'Can I contribute to the project?',
        answer: 'The project is open source and contributions, feedback and suggestions are always welcome.',
        showing: false,
    },
    {
        question: 'How long do you store data for?',
        answer: 'Data is only stored for as long as the service may be in use. Data associated with an account is scheduled for deletion after 3 months of inactivity (i.e. no access to the dashboard, monitor, or data API).',
        showing: false,
    },
    {
        question: 'How can I delete my logged request data?',
        answer: 'You can delete your account and all associated data at any time by entering your API key <a href="/delete">here</a>.',
        showing: false,
    },
    {
        question: 'How can I amend or modify my logged request data?',
        answer: "The service currently doesn't allow for the ability for users to directly amend or modify their logged request data themselves. If you have a specific request, please get in contact with us.",
        showing: false,
    },
    {
        question: 'How can I request my data?',
        answer: `You can access your raw logged request data by sending a GET request to our data API.<br><br><pre style="font-size:0.85em; text-wrap: auto; padding: 1em 0;background: var(--dark-background); padding: 1.4em 2em; font-weight: 500; border-radius: 0.3em;"><code>curl --header "X-AUTH-TOKEN: &lt;API-KEY&gt;" https://apianalytics-server.com/api/data</code></pre><br>A range of URL parameters can be used to filter the data:<br><br><ul style="list-style: circle; padding-left: 1.5em;">
  <li><code style="color: white; padding-right: 0.2em">page:</code> The page number, with a maximum page size of 50,000 (defaults to 1).</li>
  <li><code style="color: white; padding-right: 0.2em">date:</code> The exact day the requests occurred on (yyyy-mm-dd).</li>
  <li><code style="color: white; padding-right: 0.2em">dateFrom:</code> The lower bound of a date range for when the requests occurred (yyyy-mm-dd).</li>
  <li><code style="color: white; padding-right: 0.2em">dateTo:</code> The upper bound of a date range for when the requests occurred (yyyy-mm-dd).</li>
  <li><code style="color: white; padding-right: 0.2em">hostname:</code> The hostname of your service.</li>
  <li><code style="color: white; padding-right: 0.2em">ipAddress:</code> The IP address of the client.</li>
  <li><code style="color: white; padding-right: 0.2em">status:</code> The status code of the response.</li>
  <li><code style="color: white; padding-right: 0.2em">location:</code> A two-character location code of the client.</li>
  <li><code style="color: white; padding-right: 0.2em">user_id:</code> A custom user identifier (only relevant if a <code style="color: white; padding: 0 0.2em;">get_user_id</code> mapper function has been set within the config).</li>
</ul><br>Example:<br><br><pre style="font-size:0.85em; text-wrap: auto; padding: 1em 0; background: var(--dark-background); padding: 1.4em 2em; font-weight: 500; border-radius: 0.3em;"><code>curl --header "X-AUTH-TOKEN: &lt;API-KEY&gt;" https://apianalytics-server.com/api/data?page=3&dateFrom=2022-01-01&hostname=apianalytics.dev&status=200&user_id=b56cbd92-1168-4d7b-8d94-0418da207908</code></pre>`,
        showing: false,
    },
    {
        question: 'The wrong value is being logged for each request, how can I change this?',
        answer: `APIs can be deployed in a wide-range of environments, and the ideal values to be extracted from a request can differ. The API Analytics middleware is flexible in how you control this, allowing custom mapping functions to determine how a given field is extracted from a request.<br><br>For example, if your API is behind a reverse proxy, the client's IP address may be re-written to the header <code style="color: white; padding: 0 0.2em;">X-Real-IP</code>.<br><br>
         
        <pre style="font-size:0.85em; text-wrap: auto; padding: 1em 0; background: var(--dark-background); padding: 1.4em 2em; font-weight: 500; border-radius: 0.3em; overflow: auto;"><code class="block !text-sm font-normal s-XmipeXcWWjoL language-python"><span class="token keyword">from</span> fastapi <span class="token keyword">import</span> FastAPI
<span class="token keyword">from</span> api_analytics<span class="token punctuation">.</span>fastapi <span class="token keyword">import</span> Analytics<span class="token punctuation">,</span> Config

config <span class="token operator">=</span> Config<span class="token punctuation">(</span>
    get_ip_address<span class="token operator">=</span><span class="token keyword">lambda</span> request<span class="token punctuation">:</span> request<span class="token punctuation">.</span>headers<span class="token punctuation">.</span>get<span class="token punctuation">(</span><span class="token string">'X-REAL-IP'</span><span class="token punctuation">,</span> <span class="token string">''</span><span class="token punctuation">)</span>
<span class="token punctuation">)</span>

app <span class="token operator">=</span> FastAPI<span class="token punctuation">(</span><span class="token punctuation">)</span>
app<span class="token punctuation">.</span>add_middleware<span class="token punctuation">(</span>Analytics<span class="token punctuation">,</span> api_key<span class="token operator">=</span><span class="token operator">&lt;</span>API<span class="token operator">-</span>KEY<span class="token operator">&gt;</span><span class="token punctuation">,</span> config<span class="token operator">=</span>config<span class="token punctuation">)</span></code></pre>`,
        showing: false,
    }
];

export default faq;