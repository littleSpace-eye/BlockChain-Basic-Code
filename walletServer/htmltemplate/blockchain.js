$(document).ready(
  $(function () {

    document.getElementById('copyButton').addEventListener('click', function () {
      var input1Value = document.getElementById('private_key').value;
      var input2Value = document.getElementById('public_key').value;
      var input3Value = document.getElementById('blockchain_address').value;

      var formattedText = "钱包账号:\n" + "privateKey:" + input1Value + '\n' + "publicKey:" + input2Value + '\n' + "Address:" + input3Value;

      navigator.clipboard.writeText(formattedText)
        .then(function () {
          console.log('文本已成功复制到剪贴板');
          // 或者你可以显示一个提示消息给用户，表示复制成功
          alert("复制成功")
        })
        .catch(function (error) {
          console.error('复制文本到剪贴板时出错:', error);
          // 复制失败时，你可以根据需要处理错误情况
          alert("复制失败")
        });
    });


    $.ajax({
      url: "http://localhost:8080/wallet/walletGetBlockChain",
      type: "GET",
      dataType: "json",
      success: function (response) {
        var chain = response.chain;
        var container = document.getElementById("get_blockchain"); // Replace "get_blockchain" with the ID of your container element

        chain.forEach(function (item, index) {
          // Create a new button
          var button = document.createElement("button");
          button.type = "button";
          button.className = "btn  btn-primary dropdown-toggle custom-button mt-3";
          button.setAttribute("data-bs-toggle", "dropdown");
          button.setAttribute("aria-expanded", "false");
          button.innerText = "Block " + index;

          // Create a new card div
          var cardDiv = document.createElement("div");
          cardDiv.className = "card border-dark mb-3";
          cardDiv.style.maxWidth = "18rem";
          cardDiv.style.display = "none"; // Hide the card initially

          // Create the card header
          var cardHeaderDiv = document.createElement("div");
          cardHeaderDiv.className = "card-header";
          cardHeaderDiv.innerText = "Block Number: " + index;

          // Create the card body
          var cardBodyDiv = document.createElement("div");
          cardBodyDiv.className = "card-body";

          // Create the card title with the block number
          var cardTitleH5 = document.createElement("h5");
          cardTitleH5.className = "card-title";
          cardTitleH5.innerText = "Block Number: " + index;

          // Create the card text with the block content
          var cardTextP = document.createElement("p");
          cardTextP.className = "card-text";

          // Construct the string representation of the item object
          var blockContent = "Block Content:\n";
          blockContent += "Timestamp: " + item.timestamp + "\n";
          blockContent += "Nonce: " + item.nonce + "\n";
          blockContent += "Previous Hash: " + item.previous_hash + "\n";

          // Iterate over the transactions array
          blockContent += "Transactions:\n";
          item.transactions.forEach(function (transaction, index) {
            blockContent += "Transaction " + (index + 1) + ":\n";
            blockContent += "Sender: " + transaction.sender_blockchain_address + "\n";
            blockContent += "Recipient: " + transaction.recipient_blockchain_address + "\n";
            blockContent += "Value: " + transaction.value + "\n";
            blockContent += "Transaction Hash: " + transaction.transaction_hash + "\n";
          });

          // Set the inner text of the card text element
          cardTextP.innerText = blockContent;

          // Append the card title and text to the card body
          cardBodyDiv.appendChild(cardTitleH5);
          cardBodyDiv.appendChild(cardTextP);

          // Append the card header and body to the card div
          cardDiv.appendChild(cardHeaderDiv);
          cardDiv.appendChild(cardBodyDiv);

          // Create a new list item
          var listItem = document.createElement("li");
          listItem.className = "dropdown-item";
          listItem.appendChild(button);

          // Append the list item to the container
          container.appendChild(listItem);
          container.appendChild(cardDiv);

          // Event listener for button click
          button.addEventListener("click", function () {
            // Toggle the display of the card
            if (cardDiv.style.display === "none") {
              cardDiv.style.display = "block";
            } else {
              cardDiv.style.display = "none";
            }
          });
        });
      },
      error: function (response, error) {
        // Handle the error response
        console.log("Error in request:", response);
        console.error(error);
      }
    });

    $("#btn-get-block-by-num").click(function () {
      let numGetBlockValue = $("#get_block_num").val();
      $.ajax({
        url: "http://localhost:8080/wallet/walletByNumCheckBlock",
        type: "POST",
        data: {
          num: numGetBlockValue
        },
        success: function (response) {
          // Construct the string representation of the item object
          var blockContent = "Block Transaction:\n";

          // Iterate over the transactions array
          blockContent += "Transactions:\n";
          response.transactions.forEach(function (transaction, index) {
            blockContent += "Transaction " + (index + 1) + ":\n";
            blockContent += "Sender: " + transaction.sender_blockchain_address + "\n";
            blockContent += "Recipient: " + transaction.recipient_blockchain_address + "\n";
            blockContent += "Value: " + transaction.value + "\n";
            blockContent += "Transaction Hash: " + transaction.transaction_hash + "\n";
          });
          // Populate blockContent into the card elements
          $("#card-title").text("Block: " + numGetBlockValue);
          $("#card-Timestamp").text("Timestamp:"+response.timestamp);
          $("#card-Nonce").text("Nonce:"+response.nonce);
          $("#card-PreviousHash").text("PreviousHash:"+response.previous_hash);
          $("#card-Transaction").text(blockContent);

          // Show the card
          $("#card-Block").show();
        },
        error: function (error) {
          console.error(error);
        },
      });
    })

    $("#close-button").click(function() {
      $("#card-Block").fadeOut().hide();
    });
    
    
    
    $("#close-button-Trans").click(function() {
      $(".modal").fadeOut().hide();
    });


    $("#btn-get-block-by-hash").click(function () {
      let hashGetBlockValue = $("#get-block-hash").val();
      $.ajax({
        url: "http://localhost:8080/wallet/walletGetBlockByHash",
        type: "POST",
        data: {
          hashBlock: hashGetBlockValue
        },
        success: function (response) {
             // Construct the string representation of the item object
          var blockContent = "Block Transaction:\n";

          // Iterate over the transactions array
          blockContent += "Transactions:\n";
          response.transactions.forEach(function (transaction) {
            blockContent += "Sender: " + transaction.sender_blockchain_address + "\n";
            blockContent += "Recipient: " + transaction.recipient_blockchain_address + "\n";
            blockContent += "Value: " + transaction.value + "\n";
            blockContent += "Transaction Hash: " + transaction.transaction_hash + "\n";
          });
          // Populate blockContent into the card elements
          $("#card-title").text("Block: " + hashGetBlockValue);
          $("#card-Timestamp").text("Timestamp:"+response.timestamp);
          $("#card-Nonce").text("Nonce:"+response.nonce);
          $("#card-PreviousHash").text("PreviousHash:"+response.previous_hash);
          $("#card-Transaction").text(blockContent);

          // Show the card
          $(".card").show();
        },
        error: function (error) {
          console.error(error);
        },
      });
    })


    $("#btn-get-trans-by-hash").click(function () {
      let hashGetTransValue = $("#get-trans-hash").val();
      $.ajax({
        url: "http://localhost:8080/wallet/walletGetTransByHash",
        type: "POST",
        data: {
          transactionHash: hashGetTransValue
        },
        success: function (response) {
          console.log("value!!!!",response)
          console.log(response.value)
          // Populate blockContent into the card elements
          $("#card-Trans").text("Transaction: " + hashGetTransValue);response.sender_blockchain_address
          $("#card-Trans-sender-address").text("SenderAddress:"+response.sender_blockchain_address);
          $("#card-Trans-receive-address").text("ReceiveAddress:"+response.recipient_blockchain_address);
          $("#card-Trans-hash").text("TransactionHash:"+response.transaction_hash);
          $("#card-Trans-value").text("Value:"+response.value);

          //Show the card
          $(".modal").show();
        },
        error: function (error) {
          console.error(error);
        },
      });
    })




    $("#get_amount").click(function () {
      var blockaddress = $("#blockchain_address").val();
      let _postdata = {
        blockchain_address: blockaddress,
        // 添加其他参数...
      };

      console.log("blockaddress:", JSON.stringify(_postdata));

      $.ajax({
        url: "http://127.0.0.1:8080/wallet/amount",
        type: "POST",
        contentType: "application/json",
        data: JSON.stringify(_postdata),
        success: function (response) {
          $("#wallet_amount").text(response["amount"]);

          console.info(response);
        },
        error: function (error) {
          console.error(error);
        },
      });
    });

    $("#reload_wallet").click(function () {
      $.ajax({
        url: "http://127.0.0.1:8080/wallet",
        type: "POST",
        success: function (response) {
          $("#public_key").val(response["public_key"]);
          $("#private_key").val(response["private_key"]);
          $("#blockchain_address").val(response["blockchain_address"]);
          console.info(response);
        },
        error: function (error) {
          console.error(error);
        },
      });
    });


    $("#loadWalletByPrivatekey").click(function () {
      var privateKeyValue = $("#private_key").val();
      // alert("私钥的值是：" + privateKeyValue);
      $.ajax({
        url: "http://127.0.0.1:8080/walletByPrivatekey",
        type: "POST",
        data: {
          privatekey: privateKeyValue,
          // 添加其他参数...
        },
        success: function (response) {
          $("#public_key").val(response["public_key"]);
          $("#private_key").val(response["private_key"]);
          $("#blockchain_address").val(response["blockchain_address"]);
          console.info(response);
        },
        error: function (error) {
          console.error(error);
        },
      });
    });


    $("#send_money_button").click(function () {
      let confirm_text = "确定要发送吗?";
      let confirm_result = confirm(confirm_text);
      if (confirm_result !== true) {
        alert("取消");
        return;
      }

      let transaction_data = {
        sender_private_key: $("#private_key_trans").val(),
        sender_blockchain_address: $("#blockchain_address_trans").val(),
        recipient_blockchain_address: $("#recipient_address").val(),
        sender_public_key: $("#public_key_trans").val(),
        value: $("#send_amount").val(),
      };

      $.ajax({
        url: "http://localhost:8080/transaction",
        type: "POST",
        contentType: "application/json",
        data: JSON.stringify(transaction_data),
        success: function (response) {
          console.info("response:", response);
          console.info("response.message:", response.message);
          console.log(JSON.stringify(transaction_data))
          if (response.message === "fail") {
            alert("failed:余额不足");
            return;
          }

          alert("发送成功" + JSON.stringify(response));
        },
        error: function (response) {
          console.log(JSON.stringify(transaction_data))
          console.error(response);
          alert("发送失败");
        },
      });


    });
  })


)

