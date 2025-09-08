document.addEventListener("DOMContentLoaded", () => {
  const filterInput = document.getElementById("galleryFilter");
  const displayFilter = document.getElementById("displayFilter");
  const cards = document.querySelectorAll("#gallery .card-link");

  function applyFilters() {
    const term = filterInput.value.toLowerCase();
    const filterType = displayFilter.value;

    cards.forEach(card => {
      const title = card.querySelector(".title")?.textContent.toLowerCase() || "";
      const artist = card.querySelector(".name")?.textContent.toLowerCase() || "";
      const type = card.getAttribute("data-display");

      const matchesText = title.includes(term) || artist.includes(term);
      const matchesType =
        filterType === "all" ||
        (filterType === "display" && type === "display") ||
        (filterType === "biddable" && type === "biddable");

      card.style.display = matchesText && matchesType ? "" : "none";
    });
  }

  filterInput.addEventListener("input", applyFilters);
  displayFilter.addEventListener("change", applyFilters);
});
