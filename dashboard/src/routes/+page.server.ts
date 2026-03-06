import { codeToHtml } from 'shiki';
import frameworkExamples, { frameworkLanguages } from '$lib/framework';

export async function load() {
	const opts = {
		theme: 'one-dark-pro' as const,
		colorReplacements: { '#282c34': 'transparent' }
	};

	const highlighted: Record<string, { install: string; example: string }> = {};

	for (const [framework, data] of Object.entries(frameworkExamples)) {
		const lang = frameworkLanguages[framework] ?? 'text';
		highlighted[framework] = {
			install: await codeToHtml(data.install, { lang: 'text', ...opts }),
			example: await codeToHtml(data.example, { lang, ...opts })
		};
	}

	return { highlighted };
}
