const textareas = document.querySelectorAll("textarea");
const txr1 = textareas[0];
const txr2 = textareas[1];
let scHeight = txr1.scrollHeight; // Get the scroll height of the content
txr1.style.height = `${scHeight}px`; // Adjust the height to fit content

scHeight = txr2.scrollHeight; // Get the scroll height of the content
txr2.style.height = `${scHeight}px`; // Adjust the height to fit content

txr1.addEventListener("input", (e) => {
  txr1.style.height = "40px"; // Reset height to minimum
  let scHeight = e.target.scrollHeight; // Get the scroll height of the content
  txr1.style.height = `${scHeight}px`; // Adjust the height to fit content
});
txr2.addEventListener("input", (e) => {
  txr2.style.height = "200px"; // Reset height to minimum
  let scHeight = e.target.scrollHeight; // Get the scroll height of the content
  txr2.style.height = `${scHeight}px`; // Adjust the height to fit content
});
