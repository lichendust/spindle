function set_copy() {
	if (!navigator.clipboard) {
		return
	}

	const button_label = "⌗"

	let blocks = document.querySelectorAll("pre")
	blocks.forEach((block) => {
		let button = document.createElement("button")

		button.className = "copy mono"
		button.innerText = button_label
		block.appendChild(button)

		button.addEventListener("click", async () => {
			await copy_code(block);
		})
	})

	async function copy_code(block) {
		let code = block.querySelector("code");
		let text = code.innerText;

		await navigator.clipboard.writeText(text);
	}
}

set_copy()