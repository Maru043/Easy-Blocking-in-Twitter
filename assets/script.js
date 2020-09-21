"use strict";

function sendForm() {
  const conditions = parseSubmitForm();
  if (conditions === null) return false;

  fetch(`http://${location.host}/process/`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json;charset=utf-8",
    },
    body: JSON.stringify(conditions),
  }).then(function (response) {
    console.log(response);
  });
  return false;
}

function parseSubmitForm() {
  const form = document.forms.submitForm;
  const conditions = { targetScreenNames: [] };
  conditions.runMode = form.elements.runMode.value;
  if (conditions.runMode === "") return null;
  conditions.targetScreenNames = parseTextarea(
    form.elements.targetScreenNames.value
  );

  if (conditions.runMode === "unblock" || conditions.runMode === "unmute") {
    conditions.exceptFollowing = false;
    conditions.exceptFollowers = false;
  } else {
    conditions.exceptFollowing = form.elements.exceptFollowing.checked;
    conditions.exceptFollowers = form.elements.exceptFollowers.checked;
  }
  return conditions;
}

function parseTextarea(textarea) {
  const targetScreenNames = textarea.split(/[,?、\s]/).filter(function (e) {
    return e !== "";
  });
  for (let i = 0; i < targetScreenNames.length; i++) {
    targetScreenNames[i] = targetScreenNames[i].replace(/@/g, "");
  }
  return Array.from(new Set(targetScreenNames));
}

function toggleExceptConf() {
  const form = document.forms.submitForm;
  const following = form.exceptFollowing;
  const followers = form.exceptFollowers;
}
