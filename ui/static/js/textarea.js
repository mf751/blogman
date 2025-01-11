const textareas = document.querySelectorAll("textarea");
textareas[0].style.height = "40px";
textareas[1].style.height = "200px";

textareas[0].addEventListener("input", (e) => {
  textareas[0].style.height = "40px"; // Reset height to minimum
  let scHeight = e.target.scrollHeight; // Get the scroll height of the content
  textareas[0].style.height = `${scHeight}px`; // Adjust the height to fit content
});
textareas[1].addEventListener("input", (e) => {
  textareas[1].style.height = "200px"; // Reset height to minimum
  let scHeight = e.target.scrollHeight; // Get the scroll height of the content
  textareas[1].style.height = `${scHeight}px`; // Adjust the height to fit content
});
