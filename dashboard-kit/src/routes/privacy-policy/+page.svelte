<div class="info-page-container text-center leading-snug text-[#c3c3c3]">
	<h1>Privacy Policy</h1>

	<div class="text-[#707070]">Last updated: 2025-01-28</div>

	<p>
		API Analytics ("we," "us," or "our") is committed to protecting the privacy of our users. This
		Privacy Policy explains how we collect, use, and share information when you use our analytics
		service ("Service") and outlines your rights regarding your personal data.
	</p>

	<div class="important">
		<p>
			By using our Service, you acknowledge that you have read and understood this Privacy Policy.
			If you disagree with any aspect of these terms, please discontinue use of the Service.
		</p>
	</div>

	<h2>1. Information We Collect</h2>

	<p>
		Through our API middleware package, we automatically collect the following data points for each
		API request:
	</p>

	<ul>
		<li>Request method (GET, POST, PUT, DELETE, etc.)</li>
		<li>Request endpoint (API endpoint being accessed)</li>
		<li>User agent</li>
		<li>IP address (subject to privacy level settings)</li>
		<li>Request timestamp (in UTC)</li>
		<li>Response status code</li>
		<li>Response time (in milliseconds)</li>
		<li>API hostname</li>
		<li>API framework identifier</li>
	</ul>

	<h3>Personal Information</h3>

	<p>
		When using the Service, data logged by your API may include personal information such as IP
		addresses, unless you opt out of this collection.
	</p>

	<h4>Privacy Levels for IP Address Collection</h4>

	<p>
		Client IP addresses from incoming requests to your API are handled based on the configured
		privacy level in the API Analytics middleware:
	</p>
	<ul>
		<li><strong>Level 0 (Default):</strong> IP addresses are collected and stored.</li>
		<li><strong>Level 1:</strong> IP addresses are discarded immediately after use.</li>
		<li><strong>Level 2:</strong> IP addresses are never accessed or collected.</li>
	</ul>
	<p>
		You can adjust the privacy level at any time through the API Analytics middleware configuration
		settings.
	</p>

	<h4>Optional Custom User ID Field</h4>
	<p>
		An optional custom user identifier text field can be configured to store additional data related
		to the request. The data captured by this field is fully defined by you within the API Analytics
		middleware settings.
	</p>

	<h4>Responsibility for Data Submission</h4>
	<p>
		As the operator of your API, you are solely responsible for ensuring that the submission of any
		client IP addresses or other personal data to the Service complies with all applicable laws and
		regulations, including but not limited to data protection and privacy laws in your jurisdiction
		and the jurisdictions of your API's users.
	</p>
	<p>
		If collecting and submitting IP addresses is prohibited in your use case or under local laws,
		you must configure the privacy level appropriately (Level 1 or Level 2) to disable IP address
		access and storage.
	</p>

	<h2>2. How We Use Information</h2>

	<p>We use the information we collect to provide, maintain, and improve the Service.</p>

	<p>
		Other than this use, stored data is only every used to populate the dashboard on your access, or
		returned to you via the data API on request.
	</p>

	<h2>3. Sharing and Disclosure of Information</h2>

	<p>
		We do not sell to or share your personal information with third parties. However, we may
		disclose your information if required by law or in response to valid legal requests (e.g.,
		subpoenas or court orders).
	</p>

	<h2>4. Account Management and Data Access</h2>
	<p>
		Account creation and managament for the Service is handled solely through a unique API key
		(UUID). This key is required to log API requests, access your dashboard, or retrieve raw data.
		<strong>If you lose your API key, you will be unable to access your associated data.</strong> Please
		ensure you securely store your API key.
	</p>

	<h2>5. Data Retention</h2>

	<p>
		We retain your data only for as long as necessary to fulfill the purposes outlined in this
		Privacy Policy.
	</p>

	<p>
		You can delete your account and all associated data by visiting <a
			href="https://apianalytics.dev/delete">https://apianalytics.dev/delete</a
		> and entering your API key.
	</p>

	<p>
		Accounts and associated data will be deleted automatically after a period of inactivity. This is
		triggered when one of the two following conditions are met:
	</p>
	<ul>
		<li>
			Six months have passed since the last data access, via the API Analytics dashboard or data API, or
		</li>
		<li>Three months have passed since the most recent logged request.</li>
	</ul>

	<h2>6. Data Security</h2>

	<p>
		All data is stored securely on a private London-based server. We implement industry-standard
		technical and organizational measures, including encryption and secure access protocols, to
		protect your data from unauthorized access, alteration, or disclosure.
	</p>

	<p>
		While we make every effort to ensure the security of your data, no method of transmission over the
		Internet or method of electronic storage is completely secure. Therefore, we cannot guarantee its
		absolute security.
	</p>
	<p>
		We recommend that you also take precautions to protect your data, including keeping your API key
		confidential.
	</p>

	<h2>7. Changes to This Privacy Policy</h2>
	<p>
		We may update this Privacy Policy periodically to reflect changes in our practices or for legal,
		operational, or regulatory reasons. We encourage you to review this Privacy Policy periodically
		for any updates.
	</p>
</div>

<style scoped>
	h1 {
		font-size: 2em;
		font-weight: 700;
		margin: 1.5em 0 0.5em !important;
	}

	p {
		margin: 0;
		padding: 0.8em 0;
		color: #c3c3c3;
	}

	ul {
		list-style: disc; /* Default bullet points */
		padding-left: 2em;
		margin-top: 0;
	}
	li {
		padding: 0.2em 0 0.2em 0.5em;
		color: #c3c3c3;
	}

	h2 {
		font-size: 1.5em;
		font-weight: 700;
		margin: 1.5em 0 0.5em !important;
	}
	h3 {
		margin: 1.5em 0 0.5em !important;
		font-size: 1.2em;
		font-weight: 600;
		color: #c3c3c3;
	}

	h4 {
		margin: 1.5em 0 0.5em !important;
	}
	a {
		color: var(--highlight);
	}
	a:hover {
		color: white;
	}

	.important {
		background-color: rgba(0, 0, 0, 0.15);
		border-radius: 0 4px 4px 0;
		padding: 1rem;
		border-left: 4px solid var(--highlight);
		margin: 1rem 0;
	}
	.warning {
		/* background-color: #fff3cd; */
		padding: 1rem;
		border-left: 4px solid #ffc107;
		margin: 1rem 0;
	}
</style>
