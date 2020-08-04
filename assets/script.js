"use strict";

function submitForm() {
  const form = document.forms.testForm;
  const conditions = { targetScreenNames: [] };
  const targetScreenNames = form.elements.targetScreenNames.value;

  conditions.targetScreenNames = targetScreenNames
    .split(/[,、\s]/)
    .filter(function (e) {
      return e !== "";
    });
  conditions.exceptFollowing = form.elements.exceptFollowing.checked;
  conditions.exceptFollowers = form.elements.exceptFollowers.checked;
  conditions.runMode = form.elements.runMode.value;
  conditions.blockTarget = form.elements.blockTarget.checked;
  if (conditions.runMode === "") return;
  fetch(`http://${location.host}/process/`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json;charset=utf-8",
    },
    body: JSON.stringify(conditions),
  }).then(function (response) {
    console.log(response);
  });
}
