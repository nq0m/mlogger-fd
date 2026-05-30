<script>
  // OperatorSelector component placeholder — will be fully implemented by Plan 02-03
  // Read saved operator from localStorage on mount
  let operator = $state(localStorage.getItem('fdlogger_operator') || '');

  function handleInput(e) {
    operator = e.target.value;
  }

  function handleBlur() {
    localStorage.setItem('fdlogger_operator', operator);
  }
</script>

<label>
  Operator:
  <input
    type="text"
    aria-label="Operator"
    value={operator}
    oninput={handleInput}
    onblur={handleBlur}
  />
</label>
