## V0:
1.setting up environment to store secret(will use secret manager or hashicorp or unix password manager for v1)
<br>
2.simple ui with a text box that once a mail id is typed will validate and then send a entry mail to that link
<br>
3. rate limit 10 min for the same mail id to send the mail again and in next version somehow late limit the person from sending mail at all for half an hour
<br>
4. the link will be valid for 24 hrs max
<br>
5. once the link is clicked he can go ahead typing a mail id, no of days the msg to be sent(max 7 days).
   <br>
6. msg can be typed or chosen from the template. for next version the templates come from gpt
   <br>
   7. add git ignore
   <br>
8. simple middleware
   <br>
  9. add panic handler everywhere
   <br>
   10. Need A ui for error popup
   <br>
   11. need to handle maximum gmails which can be sent on a day. if it exeeds the limit then we shouldnt allow them to send mail as it is paid feature. same with ses(be under the daily limit)
   <br>
   12. UUID might loose randomness. so need to handle it with date
      <br>
   13. add panic handlers
   <br>
   14. hard delete record one month post expiry needs to be implemented
      <br>
   15. every user will receive a unique link that will route to the same ui page. only with that link. solve this using jwt token

    <br>
   16. help me in generating a landing page for good Morning sunshine. the page should have a button of a hen to transition between both dark and light mode and a cute animation to transition from dark and light mode. the dark mode's there is grey background with white text and the light mode's color palatte is white background with black text. the main important feature is a email text area in which they have to fill the email to get started with using good morning sunshine. Once a send email button is clicked a pop up needs to come stating that there email is sent successfully go check your email and spam folder. good morning sunshine is a website which sends you cute and fancy good morning or good night message based on their selection for a fixed number of days. the website should have a sleak nav bar and a footer with copyright, developer's info links with icon's
