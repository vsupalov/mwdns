// JavaScript Document $(document).ready(function(){	 var zalando = new Array();	 var x = 0;	 var y = 0;	$.ajax({           url: 'http://www.zalando.de/schuhe/',           type: 'GET',		   dataType:'html',           success: function(res) {             $(res.responseText).find('.gItem .productBox img').each(function(index) {				 if (index > 9){					 return false;				 }    			 zalando[index] = $(this).attr('longdesc');				 x = x+1;			 });           }         });		 	$("#card_1").html($("#cardBackContent").html());	$("#card_2").html($("#cardContent").html());			$(".flip_card").click(function(){		 var flippingCard = $(this);    	 $(this).flippy({			content:$("#cardBackContent"),			direction:"LEFT",			duration:"350",			onStart:function(){				flippingCard.removeClass('card-boarder')				$("#cardContent img").attr('src',zalando[y]);				if (y <= x){					y = y + 1;				}			},			onFinish:function(){				flippingCard.addClass('card-boarder');				flippingCard.animate({    				//left: '-=200',  					}, 500, function() {    				// Animation complete.  				});											}		}); 	});});