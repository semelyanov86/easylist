$(function() {
    var getUrl = function(event, params) {
        return (
            (typeof event.target.href == "undefined"
                ? params["url"]
                : event.target.href) +
            "&filter=" +
            encodeURIComponent(filters.getPageFilter().serializedFilter)
        );
    };

    $(document).on({
        Redirect: function(event, params) {
            var url, href;
            if (event.ctrlKey || event.button == 1 || event.button == 4) return;

            if (
                typeof params === "undefined" ||
                (typeof params["showPreloader"] !== "undefined" &&
                    params["showPreloader"].toString().toLowerCase() === "true")
            )
                Preloader.show();

            if (
                typeof params !== "undefined" &&
                typeof event.target.href === "undefined"
            )
                url = params["url"];
            else {
                url = event.target.href;
                href = $(event.target).attr("href");

                if (href == "#" || href == "") return;
            }

            setTimeout(function() {
                window.location.href = url;
            }, 0);
        },

        ShowMessages: function(event, params) {
            var url = getUrl(event, params);

            Preloader.show();
            $.ajax({
                url: url,
                cache: false,
                dataType: "html",
                success: function(data) {
                    MessagesHelper.showInfo(data);
                    Preloader.hide();
                },
                error: function(xhr, status, error) {
                    console.log(
                        status + ":" + error + "(" + xhr.responseText + ")"
                    );
                    alert("Рљ СЃРѕР¶Р°Р»РµРЅРёСЋ, РІРѕР·РЅРёРєР»Р° РѕС€РёР±РєР°.");
                    Preloader.hide();
                }
            });
        },

        TableDataUpdate: function(event, params) {
            window.location.href = params.url;
        },

        TableDataNext: function(event) {
            var element = $(event.target);
            var params = element.tablecontrol("getParams");

            Preloader.show(null, $(".table-data-overlay")[0]);
            $.ajax({
                url: params.pageurl,
                cache: false,
                dataType: "json",
                data: { startIndex: params.quantity },
                success: function(data) {
                    var lastTr = $(params.table)
                            .children("tbody")
                            .children("tr:last"),
                        checkboxes = null;
                    lastTr.after(data.Rows.content);
                    var dataTr = $(lastTr).nextAll("tr");

                    initializeReact(dataTr);

                    lastTr.next().addClass("row_separator");
                    if (!data.HasMoreItems) {
                        element.prop("disabled", true);
                    }
                    element.trigger("UpdateCellWidth");
                    Preloader.hide();

                    if ($("#IsPortionSelected").val() == "true") {
                        checkboxes = $("#TableData").find(
                            ".table-data-checkbox"
                        );
                        checkboxes.attr("disabled", true);
                        checkboxes
                            .parents("td")
                            .attr(
                                "title",
                                "Р‘СѓРґСѓС‚ РІС‹Р±СЂР°РЅС‹ СЃР»СѓС‡Р°Р№РЅС‹Рµ РїРѕС‚СЂРµР±РёС‚РµР»Рё"
                            );
                    }
                    if ($("#SelectorSelectAll").val() == "true") {
                        checkboxes = $("#TableData").find(
                            ".table-data-checkbox"
                        );
                        checkboxes.prop("checked", true);
                        checkboxes.parents("tr").addClass("row_selected");
                    }

                    dataTr.find(".switch").each(function() {
                        RestyleSwitchButton(this);
                    });

                    element.trigger("TableDataNextLoaded");
                },
                error: function(xhr, status, error) {
                    console.log(
                        status + ":" + error + "(" + xhr.responseText + ")"
                    );
                    alert("Рљ СЃРѕР¶Р°Р»РµРЅРёСЋ, РІРѕР·РЅРёРєР»Р° РѕС€РёР±РєР°.");
                    Preloader.hide();
                }
            });
        },

        UpdateDataChecked: function(event, params) {
            var tableSelectedItemsCount = $("#table_selected_items_count");
            var selecteItemsFilterStats = $("#selecteItemsFilterStats");

            if ($("#SelectorSelectAll").val() == "true") {
                tableSelectedItemsCount.text(
                    $(".filtered-accessible-count-container")
                        .first()
                        .text() + " (РІСЃРµ)"
                );
                selecteItemsFilterStats
                    .removeClass("filter__stats_checked-part")
                    .addClass("filter__stats_checked-all")
                    .show();
                return;
            }
            if ($("#IsPortionSelected").val() == "true") {
                tableSelectedItemsCount.text(
                    $("#PortionSize").val() + " (СЃР»СѓС‡Р°Р№РЅС‹С…)"
                );
                selecteItemsFilterStats
                    .removeClass("filter__stats_checked-all")
                    .addClass("filter__stats_checked-part")
                    .show();
                return;
            }
            var checkboxes = $("#TableData").find(
                ".table-data-checkbox:checked"
            );
            if (checkboxes.length > 0) {
                tableSelectedItemsCount.text(checkboxes.length);
                selecteItemsFilterStats
                    .removeClass("filter__stats_checked-all")
                    .addClass("filter__stats_checked-part")
                    .show();
                return;
            }
        },

        TableDataChecked: function(event, params) {
            var checkboxes = $("#TableData").find(".table-data-checkbox"),
                selectorSelectAll = $("#SelectorSelectAll"),
                isPortionSelected = $("#IsPortionSelected"),
                tableSelectedItemsCount = $("#table_selected_items_count"),
                selecteItemsFilterStats = $("#selecteItemsFilterStats"),
                tableDataDropdownChecked = $("#TableDataDropdownChecked");

            switch (params) {
                case "all":
                    selectorSelectAll.val(true);
                    isPortionSelected.val(false);
                    checkboxes.prop("checked", true);
                    checkboxes.parents("tr").addClass("row_selected");
                    checkboxes.removeAttr("disabled");
                    checkboxes.parents("td").removeAttr("title");
                    tableSelectedItemsCount.text(
                        $("span.filtered-accessible-count-container")
                            .first()
                            .text() + " (РІСЃРµ)"
                    );
                    selecteItemsFilterStats
                        .removeClass("filter__stats_checked-part")
                        .addClass("filter__stats_checked-all")
                        .fadeIn(600);
                    UpdateClassForDropdownCheckbox(false, true);
                    break;
                case "none":
                    selectorSelectAll.val(false);
                    isPortionSelected.val(false);
                    checkboxes.prop("checked", false);
                    checkboxes.parents("tr").removeClass("row_selected");
                    checkboxes.removeAttr("disabled");
                    checkboxes.parents("td").removeAttr("title");
                    tableSelectedItemsCount.text(0);
                    selecteItemsFilterStats
                        .fadeOut(300)
                        .removeClass(
                            "filter__stats_checked-part filter__stats_checked-all"
                        );
                    UpdateClassForDropdownCheckbox(false, false);
                    break;
                case "part":
                    // РѕР±РЅРѕРІР»СЏРµС‚СЃСЏ РїРѕСЃР»Рµ РїРѕРїР°РїР° "Р’С‹Р±СЂР°С‚СЊ С‡Р°СЃС‚СЊ" РІ function selectPortion(portionSize)
                    break;
                case "onpage":
                    selectorSelectAll.val(false);
                    isPortionSelected.val(false);
                    checkboxes.prop("checked", true);
                    checkboxes.parents("tr").addClass("row_selected");
                    checkboxes.removeAttr("disabled");
                    checkboxes.parents("td").removeAttr("title");
                    tableSelectedItemsCount.text(checkboxes.length);
                    selecteItemsFilterStats
                        .removeClass("filter__stats_checked-all")
                        .addClass("filter__stats_checked-part")
                        .fadeIn(600);
                    UpdateClassForDropdownCheckbox(true, false);
            }

            $(document).trigger("OnTableDataChecked", {
                mode: params,
                count: checkboxes.length
            });
        },

        TableRowShow: function(event, params) {
            function updateDataAndShow(data) {
                var rowNext = document.createElement("tr");
                rowNext.innerHTML = data;
                element[0].parentNode.insertBefore(
                    rowNext,
                    element[0].nextSibling
                );
                element[0].className += " row_expand";
                var cell = rowNext.children;
                for (var i = 0; i < cell.length; i++) {
                    if (cell[i].children.length === 0) continue;
                    cell[i].innerHTML =
                        '<div style="display:none;">' +
                        cell[i].innerHTML +
                        "</div>";
                    cellData.push(cell[i].children[0]);
                }
                rowNext.className = "row_expand row_expand-content";
                $(cellData).slideDown({
                    duration: 300,
                    start: function() {
                        $(window).trigger("UpdateCellWidth");
                    },
                    complete: function() {
                        element.data("isProcessed", false);
                    }
                });
                element.data({ rowNext: rowNext, cellData: cellData });
                $(rowNext)
                    .find(".switch")
                    .each(function() {
                        RestyleSwitchButton(this);
                    });
                initializeContext(element[0].nextSibling);
                if (preloader) {
                    Preloader.hide(preloader[0]);
                }
                if (params.onsuccess) {
                    eval(params.onsuccess);
                }
            }

            var element = $(event.target),
                rowNext,
                cellData = [],
                url = typeof params === "object" ? params.url : params,
                preloader = params
                    ? params.preloader
                        ? element.find(params.preloader)
                        : false
                    : false;

            if (element.data("isProcessed")) return;

            if (element.data("init") || typeof url === "undefined") {
                element.data("isProcessed", true);
                cellData = element.data("cellData");
                rowNext = element.data("rowNext");

                if (!rowNext) rowNext = element.next("tr")[0];

                if (!cellData && rowNext)
                    cellData = $(rowNext)
                        .children()
                        .children();

                if (element.data("init") === "on") {
                    $(cellData).slideUp({
                        duration: 300,
                        start: function() {
                            $(window).trigger("UpdateCellWidth");
                        },
                        complete: function() {
                            element[0].className = element[0].className.replace(
                                " row_expand",
                                ""
                            );
                            rowNext.className = "row_expand-content";
                            element.data("isProcessed", false);
                            element.data("init", "off");
                        }
                    });
                } else {
                    if (typeof params !== "undefined" && params.deniCache) {
                        cellData = [];
                        rowNext.parentNode.removeChild(rowNext);
                        if (preloader) {
                            Preloader.show(preloader[0]);
                        }
                        $.ajax({
                            url: url,
                            cache: false,
                            success: function(data) {
                                updateDataAndShow(data);
                                element.data("init", "on");
                            },
                            error: function(xhr, status, error) {
                                if (preloader) {
                                    Preloader.hide(preloader[0]);
                                }
                            }
                        });
                    } else {
                        element[0].className += " row_expand";
                        rowNext.className = "row_expand row_expand-content";
                        $(cellData)
                            .css("display", "none")
                            .slideDown({
                                duration: 300,
                                start: function() {
                                    $(window).trigger("UpdateCellWidth");
                                },
                                complete: function() {
                                    element.data("isProcessed", false);
                                    element.data("init", "on");
                                }
                            });
                    }
                }
            } else {
                element.data("isProcessed", true);
                element.data("init", "on");
                if (preloader) {
                    Preloader.show(preloader[0]);
                }
                $.ajax({
                    url: url,
                    cache: false,
                    success: function(data) {
                        updateDataAndShow(data);
                    },
                    error: function(xhr, status, error) {
                        if (preloader) {
                            Preloader.hide(preloader[0]);
                        }
                    }
                });
            }
        },

        ToggleElements: function(event, params) {
            var element = event.target;
            var toggles =
                typeof params == "object" ? $(params.selector) : $(params);
            var callbackComplete =
                typeof params == "object" && params.callback
                    ? params.callback
                    : false;
            var callbackStep =
                typeof params == "object" && params.callbackStep
                    ? params.callbackStep
                    : false;
            var callbackParams = callbackComplete
                ? params.callbackParams
                : false;
            var linkText;

            var callbackInit = function(callbackName, callbackParams) {
                if (typeof window[callbackName] == "function") {
                    window[callbackName](callbackParams);
                } else {
                    toggles.trigger(callbackName);
                }
            };

            if ((linkText = element.getAttribute("data-event-link-text"))) {
                var oldText = element.innerHTML;
                element.setAttribute("data-event-link-text", oldText);
                element.innerHTML = linkText;
            }

            if (callbackStep) {
                toggles.slideToggle({
                    duration: 300,
                    step: function() {
                        callbackInit(callbackStep, callbackParams);
                    },
                    complete: function() {
                        if (callbackComplete)
                            callbackInit(callbackComplete, callbackParams);
                    }
                });
            } else {
                toggles.slideToggle(300, function() {
                    if (callbackComplete)
                        callbackInit(callbackComplete, callbackParams);
                });
            }
        }
    });

    $(document).on("click", ".table-data-checkbox", function() {
        $(this).tablecontrol("selectedCheck");
        FillSelectedCount();
    });

    $(document).on("click", ".js-redirect", function(event) {
        event.stopPropagation();
        var url = this.getAttribute("data-url");
        if (url) {
            window.location.href = url;
        }
    });
});

function FillSelectedCount() {
    var count = $(".table-data-checkbox:checked").length;
    var TableDataDropdownChecked = $("#TableDataDropdownChecked");

    if (TableDataDropdownChecked.length > 0) {
        TableDataDropdownChecked.dropdown(
            "destroy",
            TableDataDropdownChecked
        ).dropdown();

        if (count > 0) {
            $("#selecteItemsFilterStats")
                .removeClass("filter__stats_checked-all")
                .addClass("filter__stats_checked-part")
                .fadeIn(600);

            UpdateClassForDropdownCheckbox(true, false);
        } else {
            $("#selecteItemsFilterStats")
                .fadeOut(300)
                .removeClass(
                    "filter__stats_checked-part filter__stats_checked-all"
                );

            UpdateClassForDropdownCheckbox(false, false);
        }
    }

    $(document).trigger("OnFillSelectedCount", count);

    $("#table_selected_items_count").text(count);
    $("#SelectorSelectAll").val(false);
    $("#IsPortionSelected").val(false);
}

function UpdateClassForDropdownCheckbox(isPart, isAll) {
    var partClassName = "dropdown_checked-part";
    var allClassName = "dropdown_checked-all";
    var TableDataDropdownChecked = $("#TableDataDropdownChecked");

    if (isPart && !TableDataDropdownChecked.hasClass(partClassName)) {
        TableDataDropdownChecked.addClass(partClassName);
    } else if (!isPart) {
        TableDataDropdownChecked.removeClass(partClassName);
    }

    if (isAll && !TableDataDropdownChecked.hasClass(allClassName)) {
        TableDataDropdownChecked.addClass(allClassName);
    } else if (!isAll) {
        TableDataDropdownChecked.removeClass(allClassName);
    }
}