<?php 
include 'config.php';

    if (empty($ip) || empty($domen)) {
        header("Location: https://preview.openpanel.org/");
        exit;
    }
?>

    <div id="messageContainer" class="message" style="display: none;"></div>
    <script>
        if (window.location.hash === '#expired') {
            document.getElementById('messageContainer').innerText = 'Requested proxy has expired';
            document.getElementById('messageContainer').style.display = 'block';
        }
    </script>
Temporary link for testing <?php echo $domen;?> from server IP: <?php echo $ip;?> is being generated, <br><span>please wait.</span>


<script>
function checkResponse() {
  const url = window.location.href;
  const subdomain = url.split('domains/')[1].split('/check')[0];
  //console.log(subdomain);
  const newUrl = `https://${subdomain}.openpanel.org`;
  
  console.log(newUrl);
    fetch(newUrl)
      .then(response => response.text())
      .then(data => {
          window.location.href = `https://${subdomain}.openpanel.org`;
        }
      })
      .catch(error => {
        console.error('Error:', error);
      });
  }

</script>
