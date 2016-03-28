function mkname(variantnumber, argumentnumber) {
   	return "v" + variantnumber + "a" + argumentnumber
}

function incVarpanelCount() {
		var varpanelcount = parseInt($("#maxvarpanelcount").attr("value")) + 1
		$("#maxvarpanelcount").attr("value",varpanelcount)
		return varpanelcount
}

function decVarpanelCount() {
		var varpanelcount = parseInt($("#maxvarpanelcount").attr("value")) - 1
		$("#maxvarpanelcount").attr("value",varpanelcount)
		return varpanelcount
}

function addVariant(obj) {
		varpanelcount = incVarpanelCount()
		var newid = "varpanel" + varpanelcount
		var varpanelselector = "#" + newid
		$("#varpanel").clone().attr("id",newid).show().appendTo( "#varpanelhere" );
		// Add variant number for the user
		$(varpanelselector + "* .varlabel").append( " " + varpanelcount)
		$(varpanelselector + "* .vardescription").attr( "id","editor" + varpanelcount)
		$(varpanelselector + "* .vardescriptiontoolbar").attr( "id","vardescriptiontoolbar" + varpanelcount)
		// Add variant number for the form
		$(varpanelselector + "* .form-control").each(function(elt) {
				var curname = $(this).attr("name")
				$(this).attr("name",curname + varpanelcount)
		})
		$(varpanelselector + "* .removevariant").data("variantnumber",varpanelcount).click(function(){
			decVarpanelCount()
			var num = $(this).data("variantnumber")
			$("#varpanel" + num).remove()
		})
		new wysihtml5.Editor('editor'+ varpanelcount, {
   			toolbar: 'vardescriptiontoolbar' + varpanelcount,
   			parserRules:  wysihtml5ParserRules
			 });
		$(varpanelselector).data("argcount",0)
		$(varpanelselector + "* .argcount").attr("name","argcount" + varpanelcount)
		$(varpanelselector).data("variant",varpanelcount)
		$(varpanelselector + "* .addargument").each(function(){
			$(this).data("variant",varpanelcount)
			$(this).click(addArgumentEv);
		})
		return $(varpanelselector)
}

function setVariantNameDescription(variant, name, description) {
	variant.find(".nameinput").attr("value",name)
	variant.find(".vardescription").val(description)
}

function addArgumentToVariant(variant, name, optional, type ){
	var variantnumber = variant.data("variant")
	var vps = addArgument(variantnumber)
	var argcount = $(vps).data("argcount")
	$("#" +  mkname(variantnumber,argcount)).each(function() {
		$(this).find("td:nth-child(1) input").attr("value",name)
		if (optional) {
			$(this).find("td:nth-child(2) input").attr("checked","checked")
		}
		$(this).find("td:nth-child(3) select option[value='" + type + "']").attr("selected",true)
	});
}
function addArgumentEv(event) {
	// a number
	addArgument($(this).data('variant'))
}

function removerow(variantnumber,argumentnumber) {
	var variantid = "varpanel" + variantnumber
	var varpanelselector = "#" + variantid
	var argcount = $(varpanelselector).data('argcount')
	argcount -= 1
	$(varpanelselector).data('argcount',argcount)
	$(varpanelselector +  " * .argcount").val(argcount)
	var varargname =  mkname(variantnumber,argumentnumber)
	$("#" + varargname ).remove()
}

function addArgument(variantnumber) {
	var varargname
	var variantid = "varpanel" + variantnumber
	var varpanelselector = "#" + variantid
	var argcount = $(varpanelselector).data('argcount')
	argcount += 1
	$(varpanelselector +  " * .argcount").val(argcount)
	$(varpanelselector).data('argcount', argcount)
	varargname = mkname(variantnumber,argcount)
	$(varpanelselector +  " * tbody").append('<tr id="' + varargname + '"><td><input class="form-control" type="text" value="" name="' + varargname + 'name"></td><td><input name="'  + varargname + 'optional"  class="form-control" type="checkbox"></td><td><select class="form-control" name="' + varargname +  'type"><option value="1">{...}</option><option value="2">{...,...,...}</option><option  value="3">[...]</option><option value="4">[...,...,...]</option><option value="5">to dimen | spread dimen</option><option value="6">[..=..,..=..,..=..]</option></select></td><td><a class="removerow" href="' + varpanelselector + '">X</a></td></tr>');
	$("#" + varargname + " * .removerow").click(function() { removerow(variantnumber,argcount) })
	return varpanelselector
}

function getPanelCount(selector) {
	var panelcount = parseInt($(arg).attr("value"))
	return panelcount
}

function incPanelCount(arg) {
	var panelcount = parseInt($(arg).attr("value")) + 1
	$(arg).attr("value",panelcount)
	return panelcount
}

function decPanelCount(selector) {
	var panelcount = parseInt($(selector).attr("value")) - 1
	$(selector).attr("value",panelcount)
	return panelcount
}

function addclassoptiongroup(obj,shortdescription) {
	var panelcount = incPanelCount("#panelcount")
	var newid = "panel" + panelcount
	var newselector = "#" + newid
	$("#panel").clone().attr("id",newid).show().appendTo( "#panelhere" );
	$(newselector + " .ogshortdescription").attr("value",shortdescription)
	$(newselector + " .ogshortdescription").attr("name","optiongroup" + panelcount + "shortdescription")
	$(".addoptiontogroup").click(function() {  addclassoption(null, panelcount, "","",false) })
	$(newselector + " .removeog").click(function() {  removeoptiongroup(panelcount) })
	return panelcount
}

function addclassoption(obj,optiongroup,name,shortdescription,dflt) {
	if (dflt) {
		dflt = 'checked="checked"'
	} else {
		dflt = ''
	}
	if (optiongroup == null) {
		optiongroup = getPanelCount("#panelcount")
	}

	var panelselector = "#panel" + optiongroup
	var argumentnumber = incPanelCount(panelselector + " .argumentcount")
	$(panelselector + " .argumentcount").attr("name","panel" + optiongroup + "argumentcount")
	var cls = "argument" + argumentnumber
	var name1 = "og" + optiongroup + "optionname" + argumentnumber
	var name2 = "og" + optiongroup + "optional" + argumentnumber
	var name3 = "og" + optiongroup + "shortdesc" + argumentnumber
	$(panelselector +  " * tbody").append('<tr class="' + cls + '"><td><input class="form-control" type="text" value="' + name + '" name="' +  name1 + '"></td><td><input name="' + name2  + '" class="form-control" type="checkbox"' + dflt + '></td><td><input class="form-control" name="' + name3  + '" type="text" value="' +  shortdescription + '"></td><td><a class="removerow" href="#">X</a></td></tr>');
	$("." + cls + " .removerow").click(function() {  removeoption(optiongroup,argumentnumber) })
}

function removeoptiongroup(optiongroup) {
	var panelselector = "#panel" + optiongroup
	$(panelselector).remove()
}

function removeoption(optiongroup,argumentnumber) {
	var panelselector = "#panel" + optiongroup
	var argumentselector = " .argument" + argumentnumber
	$(panelselector + argumentselector).remove()
}

function removepackageoption(packageoptionnumber) {
	var panelselector = "#panel" + packageoptionnumber
	$(panelselector).remove()
}

function addpackageoption(name,description, dflt) {
	var panelcount = incPanelCount("#panelcount")
	var newid = "panel" + panelcount
	var newselector = "#" + newid
	$("#panel").clone().attr("id",newid).show().appendTo( "#panelhere" );

	$(newselector + " .packageoptionname").attr("value",name)
	$(newselector + " .packageoptionname").attr("name","packageoption" + panelcount + "name")

	$(newselector + " .packageoptionshortdesc").attr("value",description)
	$(newselector + " .packageoptionshortdesc").attr("name","packageoption" + panelcount + "shortdescription")

	$(newselector + " .packageoptiondefault").attr("name","packageoption" + panelcount + "default")
	if (dflt) {
		$(newselector + " .packageoptiondefault").attr("checked","checked")
	}
	$(newselector + " .removepackageoption").click(function() { removepackageoption(panelcount)  })

}

