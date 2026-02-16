// palm docs — sidebar active state
(function() {
  var path = window.location.pathname;
  var links = document.querySelectorAll('.sidebar a');
  links.forEach(function(a) {
    if (a.getAttribute('href') && path.indexOf(a.getAttribute('href').replace('..','')) !== -1) {
      a.classList.add('active');
    }
  });

  // Hash-based active state for single-page sections
  function updateHash() {
    var hash = window.location.hash;
    if (!hash) return;
    links.forEach(function(a) {
      var href = a.getAttribute('href');
      if (href && href.startsWith('#')) {
        a.classList.toggle('active', href === hash);
      }
    });
  }
  window.addEventListener('hashchange', updateHash);
  updateHash();
})();
