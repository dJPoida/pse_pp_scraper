const searchInput = document.getElementById("search")
const searchHistory = document.getElementById("searchHistory")
const searchHistoryBody = searchHistory.querySelector('tbody')
const quickLinks = document.getElementById('quickSearchSelections').querySelectorAll('a');
let searchId = 0;

String.prototype.toProperCase = function () {
    return this.split(' ')
        .map(w => w[0].toUpperCase() + w.substr(1).toLowerCase())
        .join(' ')
};

/**
 * @description call this to begin the search for a suburb
 */
addSearchJob = function(searchText) {
    searchId++;

    var rowHTML = '<td class="search-item-text">' + searchText + '</td><td class="search-item-site-count"></td><td class="search-item-page-count"></td><td class="search-item-price-count"></td><td class="search-item-price"><div class="loader"></div></td>';

    var newRow = document.createElement('tr');
    newRow.innerHTML = rowHTML;
    newRow.setAttribute("class", "search-item");
    newRow.setAttribute("data-search-id", searchId);
    searchHistoryBody.prepend(newRow);

    //Make sure the search history is visible
    searchHistory.classList.add('visible');

    var errorHandler = function(statusCode) {
        console.log("failed with status", statusCode);
        updateSearchHistoryItem(searchId, 0, 0, 0, 0);
    }
    doSearch(searchId, searchText).then(function(response) {onSearchResponse(response);}, errorHandler);
}


/**
 * @description Perform a search for the average property price in a suburb
 * @param {number} searchId 
 * @param {string} searchText 
 */
function doSearch(searchId, searchText) {
    var promiseObj = new Promise(function(resolve, reject) {
        var xhr = new XMLHttpRequest();
        var data = {
            searchId: searchId,
            searchText: searchText
        }
        xhr.open("POST", "search", true);
        xhr.setRequestHeader("Content-Type", "application/json");
        xhr.send(JSON.stringify(data));

        xhr.onreadystatechange = function(){
            if (xhr.readyState === 4){
                if (xhr.status === 200){
                    resolve(JSON.parse(xhr.response));
                } else {
                    reject(xhr.status);
                }
            }
        }   
    });
    return promiseObj;
}


/**
 * @description Update one of the previous search history items in the table
 * @param {number} searchId the original ID of the search to update
 * @param {number} priceCount the number of properties used to calculate the average price
 * @param {number} averagePrice the average price of the properties found
 * @param {string} [suburb=undefined] the evaluated suburb from the search text
 * @param {string} [state=undefined] the evaluated state from the search text
 * @param {string} [postCode=undefined] the evaluated postCode from the search text
 */
updateSearchHistoryItem = function(searchId, siteCount, pageCount, priceCount, averagePrice, suburb, state, postCode) {
    var itemRow = document.querySelector('tr.search-item[data-search-id="' + searchId + '"]');
    var itemSearchTextCell = itemRow.querySelector('td.search-item-text');
    var itemPriceCountCell = itemRow.querySelector('td.search-item-price-count');
    var itemSiteCountCell = itemRow.querySelector('td.search-item-site-count');
    var itemPageCountCell = itemRow.querySelector('td.search-item-page-count');
    var itemPriceCell = itemRow.querySelector('td.search-item-price');

    if (suburb && state && postCode) {
        itemSearchTextCell.innerHTML = suburb.toProperCase() + ", " + state.toUpperCase() + " " + postCode
    }
    itemPriceCountCell.innerHTML = priceCount;
    itemSiteCountCell.innerHTML = siteCount;
    itemPageCountCell.innerHTML = pageCount;
    itemPriceCell.innerHTML = '$' + averagePrice.toLocaleString('AUD');
}


/**
 * @description fired when a search request returns a successful response
 * @param {json} responseData
*/
onSearchResponse = function (responseData) {
    updateSearchHistoryItem(
        responseData.request.searchId, 
        responseData.siteCount, 
        responseData.pageCount, 
        responseData.priceCount, 
        responseData.avgPrice,
        responseData.suburb,
        responseData.state,
        responseData.postCode,
    )
}


/**
 * @description all of the event bindings in one place. Nice!
 */
bindEvents = function() {
    searchInput.addEventListener('keydown', function(event) {
        if (event.key === 'Enter') {
            event.preventDefault();
            addSearchJob(searchInput.value);
            searchInput.value = '';
        }
    });
    
    for (var i = 0; i < quickLinks.length; i++) {
        quickLinks[i].addEventListener("click", function(event) {
            event.preventDefault();
            addSearchJob(this.getAttribute('data-search-text'));
        });
    }
}


/**
 * @description fired when the window is ready.
 */
window.onload = function() {
    bindEvents();
    searchInput.focus();
}