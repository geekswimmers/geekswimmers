Actually, the need to support a new region came from one of our most active
users. Her children used to swim with [ROW](https://www.rowswimming.ca) (Region
of Waterloo Swim Club), but the [Hanover Swim Club](https://www.hanoverswimclub.ca)
was closer to where they live. As a frequent user, she couldn't find relevant
information on GeekSwimmers relative to their new club, but before stop using
it, she asked if we could do something about it. We excitedly said: "Yes!"

[Huronia](https://en.wikipedia.org/wiki/Huronia_(region)) is a historical 
region in the province of Ontario, Canada. It is located above the Southwestern
Region and between lakes Huron and Simcoe. [Swim Ontario](https://www.swimontario.com)
defines it as an area that comprehends the following geographical regions: Bruce, 
Grey, Dufferin, Simcoe, Muskoka, Parry Sound, Kawartha Lakes, Peterborough, and
Halliburton. You can find all affiliated clubs on 
[Swim Ontario's website](https://www.swimontario.com/clubs/find-a-club/).

I genuinely thought it would be a matter of importing Huronia's time standards
into the database and voilà. But it was more than that. The database is 
actually mature to support the data, but the app was not ready to properly show
it. The main issue was that we didn't want to show Huronia's standards to 
Western swimmers neither Western's standards to Huronia's swimmers in the Time
Benchmark. To prevent that, we had to add an extra filter in the benchmark form.

![New Region Field](/static/images/content/welcome-huronia-region-new-field.png)

Well, the benchmark form is already big, and adding an extra field would make
it more time-consuming for everybody, but it had to be done. So, to make it
less annoying, we are now saving all fields to the session cookie (the one you
accept to use from time to time), which preserves the values of the fields
even when you leave the page and come back later. Now, you don't have to fill
in the benchmark form all the time. You just need to edit what has changed from
last time.

Today, we have two regions in Ontario, Canada, but tomorrow it can be
a new province, a country, even a continent. It just gets more complicated as
the dataset grows. So, very soon, we're going to have to create a user account 
where many of these selections will be part of the user profile, preventing
them to complete an even bigger benchmark form.

More challenges come together with a user account, such as the introduction of
terms of use, data protection rules, data accessibility, privacy, and many
more, but that's unavoidable if we want to grow beyond our neighbourhood.

Thank you for supporting our initiative! We hope GeekSwimmers continues to be
useful for you and your club.