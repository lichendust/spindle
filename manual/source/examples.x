& manual

title = Examples



# Image Gallery

So you've learned about tokens, encapsulating tokens and blocks.

Let's make an image gallery using those concepts.  It should be noted that for flexibility and clarity, Method 1 is preferred here.

## Method 1: Block

So this is a drop-in gallery component.  You can use this anywhere, safe the knowledge that it carries its version of `!` with it and you won't get shadowed templating issues.

code raw {
	[gallery] = {
		[!] = <img class="gallery-img" src="%{link %%}">

		<div class="gallery">
			. %%
		</div>
	}

	gallery {
		! image1.jpg
		! image2.jpg
		! image3.jpg
	}
}

The power of this mode is that you can also mix types of images within the same structure:

code raw {
	[gallery] = {
		[!]  = <img src="%{link %%}">
		[!!] = <img alt="%2" src="%{link %1}">

		<div class="gallery">
			. %%
		</div>
	}

	gallery {
		!  image1.jpg
		!! image2.jpg "which has alt-text"
		!  image3.jpg
	}
}

## Method 2: Encapsulation

Given that we're mostly just using a repeated list of `!` tokens, you might have noticed there's a another way.

This method doesn't allow for having different image-types in the list, but it's a much more minimalist front-end:

code raw {
	{!} = <div class="gallery">%%</div>
	[!] = <img class="gallery-img" src="%{link %%}">

	! image1.jpg
	! image2.jpg
	! image3.jpg
}

## Method 3: Array

We've already sacrificed having different versions of images, so let's go one step further:

code raw {
	[!] = {
		<div class="gallery">

		for %% {
			<img class="gallery-image" src="%{link %it}">
		}

		</div>
	}

	! image1.jpg image2.jpg image3.jpg
}