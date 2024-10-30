# To-Do List

## V0:

- [x] Setting up environment to store secret (will use secret manager or HashiCorp or Unix password manager for v1)
- [x] Simple UI with a text box that once a mail ID is typed will validate and then send an entry mail to that link
- [x] Rate limit 10 min for the same mail ID to send the mail again and in next version somehow rate limit the person from sending mail at all for half an hour
- [x] The link will be valid for 24 hrs max
- [x] Once the link is clicked, the user can go ahead typing a mail ID, number of days the message to be sent (max 7 days)
- [ ] Message can be typed or chosen from the template. For the next version, the templates come from GPT
- [x] Add `.gitignore`
- [x] Simple middleware
- [ ] Add panic handler everywhere
- [x] Need a UI for error popup
- [ ] Need to handle maximum Gmails that can be sent in a day. If it exceeds the limit, then we shouldn't allow them to send mail as it is a paid feature. Same with SES (be under the daily limit)
- [ ] UUID might lose randomness, so need to handle it with date
- [ ] Hard delete record one month post-expiry needs to be implemented
- [x] Every user will receive a unique link that will route to the same UI page. Only with that link. Solve this using JWT token. Send the initial email ID as well in the JWT so that we can ensure an email ID can use this feature at max n times in a month using rate limit middleware
  
- [x] Should be able to display the number of days remaining along with a premium feature to unlock mailing any days they want (max 30 days) + unlock AI generation
- [x] Check if the user already has a JWT token; if so, return the same token
- [x] Error UI show (red for failure, green for success)
- [ ] Need to take care of different time zones
- [x] fix night mode in main page
- [ ] validate recordid from ui
- [x] www. gms.com should redirect to /sec/home
- [ ] Activate deactivated main id's
- [ ] check if systemd automatically restarts when there is a server shutdown
- [ ] Add a favicon to the whole page 
- [x] Add timezone feature and time customization feature
- [ ]Add versioning
- [ ]automate versioning using github action
- [ ]timezone column in db along with ispremium
- [x]inmemorycache
- [x]the job should run every minute and read result from the cache. 
- [x]delete the result from the cache after expiring
- [x] cache should be reloaded interms of sudden restart
- [x]db upgrade
- [x]remove keep hour and minute alone remove second from the cache
- [ ]write a bash script makefile to make which includes compiling tailwind config minify for production and build
- [ ]js code cleanup and improve readability
- [x]display the timezone in main page and remove 6am ist

  <h1>Resources</h1>
-  https://iconsvg.xyz/
-  https://www.material-tailwind.com/docs/html/button
-  https://razorpay.com/docs/payments/payment-gateway/web-integration/standard/integration-steps#1-build-integration

  
