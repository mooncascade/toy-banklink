var urlParams = new URLSearchParams(window.location.search);

async function populateBanks() {
    const response = await fetch('http://localhost:3000/api/banks');
    const datas = await response.json();
    console.log(JSON.stringify(datas));
    $.each(datas.results, function (i, data) {
        var imageHtml = "<img src='" + data.logo + "' style='width:100px;height:100px'>"
        $("#banks-list").append("<a href='#' onclick=paywithbank('" + data.id + "') class='bank list-group-item list-group-item-action border'>" + imageHtml + "</li>");
        $("#banks-list").append("<a href='#' onclick=paywithbank('" + data.id + "') class='bank list-group-item list-group-item-action border'>" + imageHtml + "</li>");
    });
}
async function populateFields() {
    const data = await getPaymentData()
    console.log(JSON.stringify(data));
    $("#payment-uuid").append(data.uuid);
    $("#receiver-id").append(data.receiver_id);
    $("#payment-amount").append("£" + (data.amount / 100).toFixed(2).replace(/\d(?=(\d{3})+\.)/g, '$&,'))
    $("#payment-currency").append("GBP");
    $("#receiver-name").append("Demo user");
    $("#receiver-account").append("10203012345678");

    var statusColor = "orange";
    if (data.status == "executed") {
        statusColor = "green";
        $('#banks-card').remove()
        if (urlParams.get("notify") == "true") {
            $('#success-modal').modal('show')
        }
    }
    else if (data.status == "cancelled") {
        statusColor = "red"
        populateBanks()
        if (urlParams.get("notify") == "true") {
            $('#cancelled-modal').modal('show')
        }
    }
    else {
        populateBanks()
    }

    $("#payment-status").append("<span style='color:" + statusColor + "'>" + data.status + "</span>");
}
document.addEventListener('DOMContentLoaded', function () {
    populateFields();
}, false);

async function getPaymentData() {
    const response = await fetch('http://localhost:3000/api/payment/' + urlParams.get("uuid"));
    return await response.json();
}

async function postPayment(data) {
    var response = await fetch("http://localhost:3000/api/pay", {
        method: 'POST',
        body: JSON.stringify(data),
        headers: {
            'Content-Type': 'application/json'
        }
    });
    const json = await response.json();
    try {
        console.log('Success:', JSON.stringify(json));
    } catch{
        console.log('FAILED: ', json)
    }
    return json.url
}
async function paywithbank(bank_id) {
    var values = await getPaymentData()
    var data = {
        uuid: urlParams.get("uuid"),
        amount: values.amount,
        currency: "GBP",
        beneficiary_name: "Demo user",
        beneficiary_reference: values.receiver_id,
        beneficiary_sort_code: "102030",
        beneficiary_account_number: "12345678",
        remitter_reference: "re reference",
        remitter_provider_id: bank_id
    }
    var url = await postPayment(data)
    console.log("url: " + url)
    window.location.replace(url)
    console.log("Paying with " + bank_id)
}
