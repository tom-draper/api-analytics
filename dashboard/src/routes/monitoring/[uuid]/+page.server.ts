import { redirect } from '@sveltejs/kit';

export function load({ params }) {
    const uuid = params.uuid;
    if (!uuid) {
        throw redirect(308, '/monitor');
    }

    throw redirect(308, `/monitor/${uuid}`);
}