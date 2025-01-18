// DOM Elements
const addRoutines = document.getElementById("routine");
const addEntry = document.getElementById("add-routine-btn");
const addDay = document.getElementById("routinedate")
const addTime = document.getElementById("addTime")
const routineList = document.getElementById("routine-List");
const routineForm = document.getElementById("routineForm")
const fetchEventsBtn = document.getElementById("fetchEventsBtn");
const fetchDateInput = document.getElementById("fetchDate");
const eventsContainer = document.getElementById("eventsContainer");

//validation for input
function validateInput() {
    const routine = addRoutines.value.trim();
    const date = addDay.value.trim();
    const time = addTime.value.trim();
    if (routine === "") {
        alert("Please enter a routine.");
        return false;
    }
    if (date === "") {
        alert("Please select a date.");
        return false;
    }
    if (time === "") {
        alert("Please select a time.");
        return false;
    }
    return true;
}
//Event Listner for Add Routine
addEntry.addEventListener("click", async(event) => {
    event.preventDefault();
    if (!validateInput()) {
        return;
    }
    const routineData = {
        date: addDay.value.trim(),
        time: addTime.value.trim(),
        text: addRoutines.value.trim()
    };

    try {
        // Send data to the backend
        const response = await addRoutine(routineData);
        if (response.message) {
            alert(response.message);
        }

        // If successful, update the UI
        createRoutineElement(routineData);

    } catch (error) {
        console.error('Error adding routine:', error);
        alert('Failed to add routine. Please try again.');
    }
});

    //function for Add Routine
    function createRoutineElement(routine) {
        console.log('Creating element for routine:',routine);
        const item = document.createElement("div");
        item.classList.add("routine", "dropDown", "setTime");

        const dateSpan = document.createElement("span");
        dateSpan.classList.add("routine-date");
        dateSpan.textContent = `${routine.date} `;
        item.appendChild(dateSpan);

        const timeSpan = document.createElement("span");
        timeSpan.classList.add("routine-time");
        timeSpan.textContent = `${routine.time} `;
        item.appendChild(timeSpan);

        // Create the content for the new routine item
        const routineText = document.createElement("span");
        routineText.classList.add("routine-text");
        routineText.textContent = `${routine.text} `;
        item.appendChild(routineText);

        const doneButton = document.createElement("button");
        doneButton.classList.add("Done");
        doneButton.style.color = "blue";
        doneButton.innerHTML = "<b>Done</b>";
        item.appendChild(doneButton);

        //add entire item to routine list
        routineList.appendChild(item);

        // Reset the input field
        addRoutines.value = "";
        addDay.value = "";
        addTime.value ="";
        // Event Listener for the Done Button
        doneButton.addEventListener("click", function () {
            item.style.textDecoration = "line-through";
        });
    }
    // Event Listener for Fetching Routines
fetchEventsBtn.addEventListener("click", async () => {
    const selectedDate = fetchDateInput.value.trim();
    if (selectedDate === "") {
        alert("Please select a date.");
        return;
    }

    try {
        // Fetch routines from the backend
        const routines = await fetchRoutines(selectedDate);

        // Display fetched routines
        displayRoutines(routines);
    } catch (error) {
        console.error('Error fetching routines:', error);
        alert('Failed to fetch routines. Please try again.');
    }
});


// Function to View Routines
function displayRoutines(routines) {
    routineList.innerHTML = ''; // to clear existing routines

    if (!Array.isArray(routines) || routines.length === 0) {
        routineList.innerHTML = `<p>No routines found for the selected date.</p>`;
        return;
    }

    routines.forEach(routine => {
        const routineData = {
        date: routine.date,
        time: routine.time,
        text: routine.text
        };
        createRoutineElement(routineData);
    });
}
// Function to Add Routine to Backend
async function addRoutine(routineData) {
    const lambdaFunctionURL = 'https://ir7vdtbdh4nyaq57zcpywanllm0qcghl.lambda-url.ap-south-1.on.aws/routines';

    const response = await fetch(lambdaFunctionURL, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        mode: 'cors',
        body: JSON.stringify(routineData)
    });

    if (!response.ok) {
        const errorResponse = await response.json();
        throw new Error(errorResponse.error || 'Failed to add routine.');
    }

    return await response.json();
}

// Function to Fetch Routines from Backend
async function fetchRoutines(selectedDate) {
    const lambdaFunctionURL = `https://ir7vdtbdh4nyaq57zcpywanllm0qcghl.lambda-url.ap-south-1.on.aws?date=${selectedDate}`; 

    const response = await fetch(lambdaFunctionURL, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        },
        mode: 'cors'
    });

    if (!response.ok) {
        const errorResponse = await response.json();
        throw new Error(errorResponse.error || 'Failed to fetch routines.');
    }

    const data = await response.json();
    console.log('Fetched data:', data);
    return data;//check again
}
document.addEventListener('DOMContentLoaded', function() {
    if (fetchEventsBtn) {
        fetchEventsBtn.addEventListener('click', async () => {
            const selectedDate = fetchDateInput.value.trim();
            if (selectedDate === "") {
                alert("Please select a date.");
                return;
            }

            try {
                const routines = await fetchRoutines(selectedDate);
                displayRoutines(routines);
            } catch (error) {
                console.error('Error fetching routines:', error);
                alert('Failed to fetch routines. Please try again.');
            }
        });
    }
});