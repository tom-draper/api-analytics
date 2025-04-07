let translation = 1.7;

export default function animate() {
	translation = -translation;
	const el1 = document.getElementById('hover-1');
	if (el1) {
		el1.style.transform = `translateY(${translation}%)`;
	}
	const el2 = document.getElementById('hover-2');
	if (el2) {
		el2.style.transform = `translateY(${-translation}%)`;
	}
}
