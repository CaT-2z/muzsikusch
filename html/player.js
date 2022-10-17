function enqueue() {
    query = document.getElementById('query').value
    fetch('/api/queue', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({query: query})
    }).then(function(response) {
        updateQueue()
        if (response.ok) {
            //Clear the search box
            document.getElementById('query').value = ''
        } else {
            alert('Error adding to queue: '+response.statusText)
        }
    })
}

function action(endpoint) {
    fetch('/api/' + endpoint).then(function(response) {
        //Refresh queue
        updateQueue()
    })
}

function updateQueue() {
        document.getElementById('queue').contentWindow.location.reload();
}
