import { codeToHtml } from 'shiki';
import faq from '$lib/faq';

const opts = {
    theme: 'one-dark-pro' as const,
    colorReplacements: { '#282c34': '#141414' },
};

export async function load() {
    const items = await Promise.all(
        faq.map(async (item) => {
            if (!item.codes?.length) return { question: item.question, answer: item.answer };

            let answer = item.answer;
            for (let i = 0; i < item.codes.length; i++) {
                const highlighted = await codeToHtml(item.codes[i].code, {
                    lang: item.codes[i].lang,
                    ...opts,
                });
                answer = answer.replace(`{{code:${i}}}`, highlighted);
            }
            return { question: item.question, answer };
        })
    );

    return { faq: items };
}
