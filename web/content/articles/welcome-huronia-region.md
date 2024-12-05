GeekSwimmers is running its second swimming season in Canada. We struggled to 
build a data model to support time standards and records. When we thought we 
were done, Swim Canada and Swim Ontario decided to review their standards for
the next Olympic cycle, forcing us to quickly adapt. The model became more 
mature and resilient to changes, but we later discovered that the app had a
lot to catch up. It wasn't ready to cover a wider area than our western region.
So, adapting to changes became our new normal.

Actually, the need to support a new region came from one of our most active
users. Her children used to swim with ROW (Region of Waterloo Swim Club), but
another club was closer to where they live, so they moved to Hanover Swim Club
this season. As a frequent user, she couldn't find relevant information in 
GeekSwimmers relative to their new region, but before stop using it, she asked
if we could do something about it. We excitedly said: "Yes!"

I genuinely thought it would be a matter of importing Huronia's time standards
into the database and voilà. But it was more than that. The database is 
actually mature to support the data, but the app was not ready to properly show
it. The main issue was that we didn't want to show Huronia's standards to 
western swimmers neither Western's standards to Huronia's swimmers in the time
benchmark. To prevent that, we had to add an extra filter in the benchmark form.

Well, the benchmark form is already big, and adding an extra field would make
it more time-consuming for everybody, but it had to be done. So, to make it
less annoying, we are now saving all fields to the session cookie (the one you
accept to use from time to time), which preserves the values of the fields
even when you leave the page and come back later. Now, you don't have to fill
in the benchmark form all the time. You just need to edit what has changed from
last time.

Today, we added an extra region in Ontario, Canada, but tomorrow it can be
a province, a country, even a continent. It just gets more complicated as the 
dataset grows. So, very soon, we're going to have to create a user account 
where many of these selections will be part of the user profile, preventing
them to complete an even bigger benchmark form.

More challenges come together with a user account, such as the introduction of
terms of use, data protection rules, data accessibility, privacy, and many
more, but that's unavoidable if we want to grow beyond our neighbourhood.

Thank you for supporting our initiative! We hope GeekSwimmers continues to be
useful for you and your club.