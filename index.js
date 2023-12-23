(async () => await WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then((result) => {
	go.run(result.instance);

	// const goHtml = getHTML()
	const goHtml = "test"
	document.getElementById("root").innerHTML = goHtml
}))();