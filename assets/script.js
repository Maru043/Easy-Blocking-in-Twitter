"use strict";

function submitForm() {
  const form = document.forms.testForm;
  const conditions = { targetScreenNames: [] };

  conditions.targetScreenNames = parseTextarea();
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

function parseTextarea() {
  const form = document.forms.testForm;
  const textarea = form.elements.targetScreenNames.value;

  const targetScreenNames = textarea.split(/[,?、\s]/).filter(function (e) {
    return e !== "";
  });
  for (let i = 0; i < targetScreenNames.length; i++) {
    targetScreenNames[i] = targetScreenNames[i].replace(/@/g, "");
  }
  return Array.from(new Set(targetScreenNames));
}
git@github.com:oreilly-japan/go-programming-blueprints.git