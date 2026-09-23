# Google PageSpeed Insights API Key

OpenPanel's **Site Manager** automatically checks Google PageSpeed for your websites daily. However, if you’re running this check on a shared hosting server with 100+ websites, Google may block access due to API request limits. If this happens, no PageSpeed data will be shown, and you may see an error like this in the network tab of Site Manager:

```
API Error: Quota exceeded for quota metric ‘Queries’ and limit ‘Queries per day’ of service ‘pagespeedonline.googleapis.com’ for consumer ‘project_number:583797351490’. Aborting due to API error.
```

To prevent this, OpenPanel users can generate a **free Google PageSpeed Insights API key** and configure it for their websites.

---

## Generating a Google PageSpeed Insights API Key

1. Open the [Google Cloud Console](https://console.cloud.google.com/apis/api/pagespeedonline.googleapis.com/) and select your project.
2. Click the **Enable** button.
3. Navigate to **Credentials > Create Credentials > API key**.
4. Wait for the *“Creating API key…”* process to finish, then copy your new API key.

---

## Adding the API Key in OpenPanel

1. Go to **OpenPanel > Site Manager > [Select a website]**.
2. Click on **"Click to add PageSpeed Insights API key"** below the PageSpeed data section.
   ![add api key](/img/panel/v2/add_api_key.png)
3. Paste the API key you copied and click **Save**.
   ![add api key](/img/panel/v2/add_key.png)

> Once saved, the API key will be used for all websites under your account.

---

## Adding a Global API Key in OpenAdmin

Administrators can configure a global API key that will apply to all users who haven’t set their own key.

1. Go to **OpenAdmin > Settings > Custom Code**.
2. Paste your API key in the **PageSpeed API Key** field.
3. Click **Save**.

![admin key](/img/panel/v2/admin_add_key.png)

---

Once configured, the API key ensures PageSpeed data is correctly retrieved for all your websites without hitting Google’s quota limits.
