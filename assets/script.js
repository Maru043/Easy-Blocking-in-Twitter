"use strict";

document
  .querySelector(".addTargetScreenName")
  .addEventListener("click", addTargetScreeenName);

function addTargetScreeenName() {
  const p = document.createElement("p");
  const label = document.createElement("label");
  const span = document.createElement("span");
  span.innerHTML = "@";
  const input = document.createElement("input");
  input.setAttribute("placeholder", "対象のTwitter IDを入力してください");
  input.setAttribute("name", "targetScreenName");
  label.appendChild(span);
  label.appendChild(input);
  p.appendChild(label);

  document.querySelector(".targetScreenNames").appendChild(p);
}

function submitForm() {
  const form = document.forms.testForm;
  const conditions = { targetScreenNames: [] };
  const targetScreenNames = form.elements.targetScreenName;
  for (let i = 0; i < targetScreenNames.length; i++) {
    if (targetScreenNames[i].value === "") {
      continue;
    }
    console.log("a");
    conditions.targetScreenNames.push(targetScreenNames[i].value);
  }
  conditions.exceptFollowing = form.elements.exceptFollowing.checked;
  conditions.exceptFollowers = form.elements.exceptFollowers.checked;
  conditions.runMode = form.elements.runMode.value;
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
