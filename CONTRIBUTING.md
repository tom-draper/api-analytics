# Contributing

Contributions, issues and feature requests are welcome.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/<your-username>/api-analytics`
3. Create a feature branch: `git checkout -b my-feature`
4. Make your changes
5. Commit: `git commit -m 'Add my feature'`
6. Push: `git push origin my-feature`
7. Open a Pull Request

## Project Structure

```
api-analytics/
├── analytics/        # Middleware packages (Python, Node.js, Go, Rust, Ruby, PHP, C#)
├── dashboard/        # SvelteKit web dashboard
├── server/           # Go server and self-hosting guide
```

## Middleware Packages

Each middleware package lives under `analytics/<language>/<framework>`. When adding or modifying a package:

- Follow the conventions of existing packages in the same language
- Ensure the middleware logs: method, path, user agent, IP address, status code, response time, hostname, and framework name
- Include a `README.md` with installation and usage instructions
- Test against the target framework before submitting

## Dashboard

The dashboard is built with SvelteKit 2 and Svelte 5. To run it locally:

```bash
cd dashboard
npm install
npm run dev
```

## Reporting Issues

Please use the [issue tracker](https://github.com/tom-draper/api-analytics/issues) and fill in the appropriate template. Include as much detail as possible - framework, package version, and steps to reproduce.

## Questions

For general questions, open a [GitHub Discussion](https://github.com/tom-draper/api-analytics/discussions) rather than an issue.
