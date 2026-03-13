import { redirect } from '@sveltejs/kit';

export function GET({ params }) {
	throw redirect(301, `/monitor/${params.uuid}`);
}
