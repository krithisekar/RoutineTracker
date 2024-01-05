const addRoutines = document.getElementById("routine");
const addEntry = document.getElementById("add-routine-btn");
const addDay = document.getElementById("daySelect")
const routineList = document.getElementById("routine-List");
function validation() {
    let val = addRoutines.value;
    if (val.trim() === "") {
        alert("Enter your Routine");
        return false;
    }
    return true;
}
addEntry.addEventListener("click", () => {
    if (validation()) {
        const routine = addRoutines.value;
        const item = document.createElement("div");
        item.classList.add("routine", "dropDown", "setTime");

        const newSelect = document.createElement("select");
        newSelect.classList.add("dropDown");
        for (let option of addDay.options) {
            const newOption = new Option(option.text, option.value, option.defaultSelected, option.selected);
            newSelect.appendChild(newOption);
        }

        item.appendChild(newSelect);
        // Create the content for the new routine item
        const routineText = document.createElement("span");
        routineText.classList.add("routine-text");
        routineText.textContent = routine;
        item.appendChild(routineText);

        const selectedTime = document.createElement("span");
        selectedTime.classList.add("setTime");
        const timeValue = addTime.value;
        selectedTime.textContent = timeValue;
        item.appendChild(selectedTime);
        const doneButton = document.createElement("button");
        doneButton.classList.add("Done");
        doneButton.style.color = "blue";
        doneButton.innerHTML = "<b>Done</b>";
        item.appendChild(doneButton);
        //add entire item to routine list
        routineList.appendChild(item);
        // Reset the input field
        addRoutines.value = "";
        doneButton.addEventListener("click", function () {
            item.style.textDecoration = "line-through";
        });
        item.appendChild(doneButton);
    }
});