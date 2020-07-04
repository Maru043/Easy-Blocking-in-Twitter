"use strict";

function submitForm() {
  const form = document.forms.testForm;
  const conditions = {};
  conditions.targetScreenName = form.elements.targetScreenName.value;
  conditions.exceptFollowing = form.elements.exceptFollowing.checked;
  conditions.exceptFollowers = form.elements.exceptFollowers.checked;
  const runMode = form.elements.runMode.value;
  fetch(`http://${location.host}/${runMode}/`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json;charset=utf-8",
    },
    body: JSON.stringify(conditions),
  }).then(function (response) {
    console.log(response);
  });
}
