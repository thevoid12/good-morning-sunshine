# To-Do List

## V0:

- ~~[ ] Setting up environment to store secret (will use secret manager or HashiCorp or Unix password manager for v1)~~
- ~~[ ] Simple UI with a text box that once a mail ID is typed will validate and then send an entry mail to that link~~
- [ ] Rate limit 10 min for the same mail ID to send the mail again and in next version somehow rate limit the person from sending mail at all for half an hour
- ~~[ ] The link will be valid for 24 hrs max~~
- ~~[ ] Once the link is clicked, the user can go ahead typing a mail ID, number of days the message to be sent (max 7 days)~~
- [ ] Message can be typed or chosen from the template. For the next version, the templates come from GPT
- ~~[ ] Add `.gitignore`~~
- ~~[ ] Simple middleware~~
- [ ] Add panic handler everywhere
- [ ] Need a UI for error popup
- [ ] Need to handle maximum Gmails that can be sent in a day. If it exceeds the limit, then we shouldn't allow them to send mail as it is a paid feature. Same with SES (be under the daily limit)
- [ ] UUID might lose randomness, so need to handle it with date
- [ ] Hard delete record one month post-expiry needs to be implemented
- ~~[ ] Every user will receive a unique link that will route to the same UI page. Only with that link. Solve this using JWT token. Send the initial email ID as well in the JWT so that we can ensure an email ID can use this feature at max n times in a month using rate limit middleware~~
  
- [ ] Should be able to display the number of days remaining along with a premium feature to unlock mailing any days they want (max 30 days) + unlock AI generation
- [ ] Check if the user already has a JWT token; if so, return the same token
- [ ] Error UI show (red for failure, green for success)
- [ ] Need to take care of different time zones
- [ ] fix night mode in main page
- [ ] validate recordid from ui
- [ ] - [ ] www. gms.com should readirect to /sec/home
- [ ] Activate deactivated main id's
