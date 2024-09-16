<script lang="ts">
	import Footer from '../components/Footer.svelte';
</script>

<div class="privacy-policy-container">
	<h1>Privacy Policy</h1>

	<div>Last updated: 2024-09-16</div>

	<p>
		API Analytics ("we," "us," or "our") is committed to protecting the
		privacy of our users. This Privacy Policy explains how we collect, use,
		and share information when you use our analytics service ("Service") and
		outlines your rights regarding your personal data.
	</p>

	<p>
		By using the Service, you agree to the collection and use of information
		in accordance with this Privacy Policy. If you do not agree with the
		terms of this Privacy Policy, please do not use the Service.
	</p>

	<h2>1. Information We Collect</h2>

	<p>
		When you use our Service, through use of the API middleware package,
		your API web server will log and forward data related to each incoming
		request to our secure server that is collected and stored securely. For
		each request, this is limited to the following information:
	</p>

	<ul>
		<li>Request method (GET, POST, PUT, DELETE, etc.)</li>
		<li>Request endpoint (API endpoint being accessed)</li>
		<li>User agent</li>
		<li>IP address</li>
		<li>Timestamp of the request</li>
		<li>Response status code</li>
		<li>Response time</li>
		<li>Hostname of API</li>
		<li>API framework in use (FastAPI, Flask, Express, etc.)</li>
	</ul>

	<h3>Personal Information</h3>

	<p>
		Unless opted out, log data may contain personal information such as IP
		addresses.
	</p>

	<p>
		Client IP address collection from each incoming request to your API may
		be disabled by setting the privacy level to a value of 1 (discard IP
		address after use) or 2 (client IP is never accessed) within the API
		Analytics middleware configuration. By default, the privacy level is set
		to 0, and IP addresses will be collected and stored.
	</p>

	<p>
		An optional custom user ID field can be used to store an additional
		field of information related to a request. The functionality for data
		capture for this field is defined by the user within the API Analytics
		middleware configuration.
	</p>

	<p>
		User account creation for API Analytics is achieved through a generation
		of a unique API key. This API key is required in order to log API
		requests with the Service, access your dashboard or raw data from the
		API. Should you lose your API key, you will have no method to access
		your associated data.
	</p>

	<h2>2. How We Use Information</h2>

	<p>
		We use the information we collect to provide, maintain, and improve the
		Service.
	</p>

	<p>
		Other than this use, stored data is only every used to populate the
		dashboard on your access, or returned to you via the data API on
		request.
	</p>

	<h2>3. Sharing and Disclosure of Information</h2>

	<p>
		We do not sell to or share your personal information with third parties.
		However, we may disclose your information if required by law or in
		response to valid legal requests (e.g., subpoenas or court orders).
	</p>

	<h2>4. Data Retention</h2>

	<p>
		We retain data for as long as necessary to fulfill the purposes outlined
		in this Privacy Policy. At any point, you may delete your account and
		all associated data by going to <a href
			>https://apianalytics.dev/delete</a
		> and entering your API key.
	</p>

	<p>
		Every user account created along with all associated account data is
		scheduled to be deleted automatically after a period of inactivity. This
		is triggered when one of the two following conditions are met:
	</p>
	<ul>
		<li>
			6 months after the most recent access to the API Analytics dashboard
			or data API, OR
		</li>
		<li>3 months after the most recent logged request</li>
	</ul>

	<h2>5. Data Security</h2>
	<p>
		All data is stored securely on a private London-based server. We take
		the security of your data seriously and use appropriate technical and
		organizational measures to protect it. However, no method of
		transmission over the Internet or method of electronic storage is 100%
		secure, and we cannot guarantee its absolute security.
	</p>

	<h2>6. Changes to This Privacy Policy</h2>
	<p>
		We may update this Privacy Policy from time to time. When we make
		changes, we will revise the "Last Updated" date at the top of this page.
		We encourage you to review this Privacy Policy periodically for any
		updates.
	</p>
</div>
<Footer />

<style scoped>
	.privacy-policy-container {
		max-width: 800px;
		margin: auto;
		padding: 4em 2em;
		text-align: left;
		color: white;
		font-family: 'Noto Sans' sans-serif;
	}
	p {
		margin: 0;
		padding: 1em 0;
	}

	ul {
		list-style: disc; /* Default bullet points */
		padding-left: 2em;
		margin-top: 0;
	}
	li {
		padding: 0.2em 0 0.2em 0.5em;
	}
	h1 {
		margin: 1.2em 0 0.5em !important;
	}
	h2 {
		font-size: 1.5em;
		margin: 1.2em 0 0.5em !important;
	}
	h3 {
		margin: 2em 0 0.5em !important;
	}
	a {
		color: var(--highlight);
	}
	a:hover {
		color: white;
	}
</style>
