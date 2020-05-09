//quantities
var qtdFixos = 0;
var qtdJogos = 0;
var qtdSorteados = 0;
var qtdOutros = 0;
var maxNum = 13;

//table with generated games
var table, opt;

//a set containing only fixed numbers
var numFixosSet = new Set();
var numSorteadosSet = new Set();
var numOutrosSet = new Set();

//an array with numbers; it could also be initialized using 'new Array()'
var numbers = [];
var numbersFixed = [];
var numComp = [];

function Initialize() {
    var idJogos = document.getElementById('jogos');
    var idFixos = document.getElementById('fixos');
    var idSorteados = document.getElementById('sorteados');
    var idOutros = document.getElementById('outros');

    opt = document.createElement('option');
    opt.value = 30;
    opt.innerHTML = 30;
    idJogos.options.add(opt);
    
    //fill a selection with the number of games
    for (var i = 50; i <= 300; i += 25)
    {
        opt = document.createElement('option');
        opt.value = i;
        opt.innerHTML = i;
        idJogos.options.add(opt);
    }

    //fill a selection with the number of fixed numbers
    for (var i = 0; i <= 10; i++)
    {
        opt = document.createElement('option');
        opt.value = i;
        opt.innerHTML = i;
        idFixos.options.add(opt);
    }

    //fill a selection with the number of chosen numbers
    for (var i = 15; i <= 40; i += 5)
    {
        opt = document.createElement('option');
        opt.value = i;
        opt.innerHTML = i;
        idSorteados.options.add(opt);
    }

    //fill a selection with the other numbers
    for (var i = 1; i <= 10; i++)
    {
        opt = document.createElement('option');
        opt.value = i;
        opt.innerHTML = i;
        idOutros.options.add(opt);
    }
}

//validate the options on the form
function Validate() {
    var valJogos = document.getElementById('jogos');
    var valFixos = document.getElementById('fixos');
    var valSort = document.getElementById('sorteados');
    var valOutros = document.getElementById('outros');

    var resp = "";

    if (valJogos.selectedIndex == 0)
        resp += "Quantidade inválida de jogos.\n";

    if (valFixos.selectedIndex == 0)
        resp += "Quantidade inválida de números fixos.\n";

    if (valSort.selectedIndex == 0)
        resp += "Quantidade inválida de números que mais saem.\n";

    if (valOutros.selectedIndex == 0)
        resp += "Quantidade inválida para os demais números.\n";

    if (resp != ""){
        alert(resp);
        return false;
    }
    else return true;
}

//validate the options on the form
function ValidateJogos() {
    var resp = "";

    if (qtdFixos != numFixosSet.size)
        resp += "Faltam números fixos para selecionar!\n";

    if (qtdSorteados != numSorteadosSet.size)
        resp += "Faltam números dos que mais saem para selecionar!\n";


    if (resp != ""){
        alert(resp);
        return true;
    }
    else return false;
}

function getParams() {
    var params = {};
    var parts = window.location.href.replace(/[?&]+([^=&]+)=([^&]*)/gi, function(m, key, value) {
        params[key] = value;
    });
    return params;
}

function clearBox(elementID) {
    document.getElementById(elementID).innerHTML = "";
    if (elementID === 'numFixosSelecionados') numFixosSet.clear();
    if (elementID === 'numSorteadosSelecionados') numSorteadosSet.clear();
}

//generate a table with clickable numbers
function GenerateFixTable() {
    var paramValues = getParams();
    qtdJogos = parseInt(paramValues['numJogos']);
    qtdFixos = parseInt(paramValues['numFixos']);
    qtdSorteados = parseInt(paramValues['numSorteados']);
    qtdOutros = parseInt(paramValues['numOutros']);

    clearBox("tabelaNumeros");
    table = document.getElementById("tabelaNumeros");

    var head = document.createElement('thead');  
    for (var i = 0; i < 8; i++)
    {
        var tr = document.createElement('tr');  
        for (var j = 1; j <= 10; j++)
        {   
            var td = document.createElement('td');
            var tg = i*10 + j;
            if (tg < 10) tg = '0' + tg;
            td.innerHTML = "<button type='button' class='btn btn-danger' value='" + (i*10 + j) +
                "' onclick='javascript:GetFixedNumber(this)'>" + tg + "</button>";
            tr.appendChild(td);
        }
        head.appendChild(tr);
    }
    table.appendChild(head);
    GenerateMaisSorteadosTable();
}

//generate a table with clickable numbers
function GenerateMaisSorteadosTable() {
    clearBox("tabelaNumerosMais");
    table = document.getElementById("tabelaNumerosMais");

    var head = document.createElement('thead');  
    for (var i = 0; i < 8; i++)
    {
        var tr = document.createElement('tr');  
        for (var j = 1; j <= 10; j++)
        {   
            var td = document.createElement('td');
            var tg = i*10 + j;
            if (tg < 10) tg = '0' + tg;
            td.innerHTML = "<button type='button' class='btn btn-primary' value='" + (i*10 + j) +
                "' onclick='javascript:GetSorteadoNumber(this)'>" + tg + "</button>";
            tr.appendChild(td);
        }
        head.appendChild(tr);
    }
    table.appendChild(head);
}

//remove a number from the list
function RemoveNode(li) {
    var val = li.getElementsByTagName('span')[0].innerHTML;
    numFixosSet.delete(parseInt(val));
    numSorteadosSet.delete(parseInt(val));
    li.parentNode.removeChild(li);
}

//add a clicked number from the table
function GetFixedNumber(b) {
    //alert(a.value);
    //addElements(b.value);
    var ul = document.getElementById('numFixosSelecionados');

    var choices = ul.getElementsByTagName('span');  //an array of span items
    if (numFixosSet.has(parseInt(b.value)) || numSorteadosSet.has(parseInt(b.value))) return;
    if (qtdFixos == choices.length) return;

    var li = document.createElement('li');          //a new li element
    var labelName = document.createElement('span'); //for the name of the object
    
    var tg = b.value;
    if (tg < 10) tg = '0' + tg;
    labelName.innerHTML = tg + '   ';

    //button to remove a number
    var buttonRm = document.createElement('input');
    buttonRm.type = 'button';
    buttonRm.className = 'btn btn-default';
    buttonRm.value = 'Remove';
    buttonRm.setAttribute('onclick', 'RemoveNode(this.parentNode);');

    li.appendChild(labelName);
    li.appendChild(buttonRm);

    ul.appendChild(li);

    numFixosSet.add(parseInt(b.value)); //somehow, I needed to force it to be an int
}

//add a clicked number from the table
function GetSorteadoNumber(b) {
    var ul = document.getElementById('numSorteadosSelecionados');

    var choices = ul.getElementsByTagName('span');  //an array of span items
    if (numFixosSet.has(parseInt(b.value)) || numSorteadosSet.has(parseInt(b.value))) return;
    if (qtdSorteados == choices.length) return;

    var li = document.createElement('li');          //a new li element
    var labelName = document.createElement('span'); //for the name of the object
    
    var tg = b.value;
    if (tg < 10) tg = '0' + tg;
    labelName.innerHTML = tg + '   ';

    //button to remove a number
    var buttonRm = document.createElement('input');
    buttonRm.type = 'button';
    buttonRm.className = 'btn btn-default';
    buttonRm.value = 'Remove';
    buttonRm.setAttribute('onclick', 'RemoveNode(this.parentNode);');

    li.appendChild(labelName);
    li.appendChild(buttonRm);

    ul.appendChild(li);

    numSorteadosSet.add(parseInt(b.value)); //somehow, I needed to force it to be an int
}

//generate the final table with the games generated
function GenerateJogos() {
    if (ValidateJogos()) return false;
    //create an array with all the number NOT on the fixed numbers
    makeSetAnArray();

    table = document.getElementById("tabelaJogos");
    var head = document.createElement('thead');

    for (var i = 1; i <= qtdJogos; i++)
    {
        //if (numComp.length == numOutrosSet.size*qtdOutros) makeSetAnArray();

        var tr = document.createElement('tr');  
        //create a set for each row
        var rowSet = new Set(numbersFixed);
        //fill the table with the fixed numbers
        for (var k = 0; k < numbersFixed.length; k++)
        {
            var td = document.createElement('td');
            var tg = numbersFixed[k];
            if (tg < 10) tg = '0' + tg;
            td.innerHTML = "<button type='button' class='btn btn-primary' value='" + tg +
                "'>" + tg + "</button>";
            tr.appendChild(td);
        }

        //and here fill the table with random numbers
        for (var j = 0; j < maxNum - numbersFixed.length; j++)
        {
            var td = document.createElement('td');
            var tg = getNumber(rowSet);
            if (tg < 10) tg = '0' + tg;
            td.innerHTML = "<button type='button' class='btn btn-success' value='" + tg +
                "'>" + tg + "</button>";
            tr.appendChild(td);
        }
        head.appendChild(tr);
    }
    table.appendChild(head);
}

//put the numbers from the set into an array
function makeSetAnArray() {
    numbers.length = 0;
    numbersFixed.length = 0;

    for (var k = 1; k <= 80; k++){
        //if they are drawn numbers
        if (numSorteadosSet.has(k)){
            numbers[numbers.length] = k;
        }
        //if they are fix numbers
        else if (numFixosSet.has(k)) numbersFixed[numbersFixed.length] = k;
        //neither fix nor drawn numbers
        else if (!numFixosSet.has(k) && !numSorteadosSet.has(k)){
            numOutrosSet.add(k);
            //how many times a non-drawn number has appeared?
            var occur = numComp.filter(function(value){
                return value === k;
            }).length;

            //if none or less than the max quantity
            if (occur < qtdOutros){
                //var c = Math.floor(Math.random() * 2);
                //to choose or not to choose
                //if (c & 1){
                numbers[numbers.length] = k;
                numComp[numComp.length] = k;
                //}
            }
        }
    }
    shuffle(numbers);
}

function shuffle(array) {
    var currentIndex = array.length, temporaryValue, randomIndex;

    // While there remain elements to shuffle...
    while (0 !== currentIndex) {
        //Pick a remaining element...
        randomIndex = Math.floor(Math.random() * currentIndex);
        currentIndex -= 1;

        //And swap it with the current element.
        temporaryValue = array[currentIndex];
        array[currentIndex] = array[randomIndex];
        array[randomIndex] = temporaryValue;
    }
    return array;
}

function getNumber(rowSet) {
    if (numbers.length == 0) makeSetAnArray();
    shuffle(numbers);
    var currentIndex = numbers.length, temporaryValue, randomIndex;

    do {
        //get a random value
        randomIndex = Math.floor(Math.random() * currentIndex);
    }while (rowSet.has(numbers[randomIndex]));

    //And swap it with the last element
    temporaryValue = numbers[currentIndex-1];
    numbers[currentIndex-1] = numbers[randomIndex];
    numbers[randomIndex] = temporaryValue;

    var randomNum = numbers[currentIndex-1];
    rowSet.add(randomNum);
    numbers.pop();
    return randomNum; 
}
