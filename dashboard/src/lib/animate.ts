let translation = 1.7;

export default function animate() {
	translation = -translation;
	const el = document.getElementById('hover-1');
	el.style.transform = `translateY(${translation}%)`;
	const el2 = document.getElementById('hover-2');
	el2.style.transform = `translateY(${-translation}%)`;

	setTimeout(animate, 9000);
}
