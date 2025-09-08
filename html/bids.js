// Utility to expand or collapse all bid sections
function setAll(open) {
  const details = document.querySelectorAll('tbody.item details');

  details.forEach(d => {
    const bidsRow = d.closest('tbody.item').querySelector('.bids');
    if (!bidsRow) return;

    if (open) {
      d.setAttribute('open', '');
      bidsRow.classList.add('open');
    } else {
      d.removeAttribute('open');
      bidsRow.classList.remove('open');
    }
  });

  // Update toggle button state
  const toggleBtn = document.getElementById('toggleAll');
  if (toggleBtn) {
    if (open) {
      toggleBtn.textContent = 'Collapse all bids';
      toggleBtn.setAttribute('aria-expanded', 'true');
    } else {
      toggleBtn.textContent = 'Expand all bids';
      toggleBtn.setAttribute('aria-expanded', 'false');
    }
  }
}

document.addEventListener('DOMContentLoaded', () => {
  const toggleBtn = document.getElementById('toggleAll');
  if (!toggleBtn) return;

  // Toggle all on click
  toggleBtn.addEventListener('click', (e) => {
    e.preventDefault();
    const details = document.querySelectorAll('tbody.item details');
    const anyClosed = Array.from(details).some(d => !d.open);
    setAll(anyClosed);
  });

  // Keep individual <details> in sync with their row
  document.querySelectorAll('tbody.item details').forEach(d => {
    d.addEventListener('toggle', () => {
      const bidsRow = d.closest('tbody.item').querySelector('.bids');
      if (!bidsRow) return;
      if (d.open) {
        bidsRow.classList.add('open');
      } else {
        bidsRow.classList.remove('open');
      }
    });
  });

  // Keyboard shortcuts: E = expand all, C = collapse all
  document.addEventListener('keydown', (e) => {
    const targetTag = e.target.tagName;
    if (targetTag === 'INPUT' || targetTag === 'TEXTAREA') return; // ignore typing in forms

    if (e.key.toLowerCase() === 'e') {
      setAll(true);
    }
    if (e.key.toLowerCase() === 'c') {
      setAll(false);
    }
  });
});
