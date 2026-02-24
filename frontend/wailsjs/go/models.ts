export namespace deck {
	
	export class ScryfallData {
	    oracleText?: string;
	    typeLine?: string;
	    manaCost?: string;
	    cmc?: number;
	    imageUri?: string;
	    priceUsd?: string;
	    priceUsdFoil?: string;
	    colorIdentity?: string;
	    tags?: string[];
	
	    static createFrom(source: any = {}) {
	        return new ScryfallData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.oracleText = source["oracleText"];
	        this.typeLine = source["typeLine"];
	        this.manaCost = source["manaCost"];
	        this.cmc = source["cmc"];
	        this.imageUri = source["imageUri"];
	        this.priceUsd = source["priceUsd"];
	        this.priceUsdFoil = source["priceUsdFoil"];
	        this.colorIdentity = source["colorIdentity"];
	        this.tags = source["tags"];
	    }
	}
	export class Card {
	    quantity: number;
	    name: string;
	    setCode?: string;
	    collectorNumber?: string;
	    foil?: boolean;
	    tags?: string[];
	    scryFall?: ScryfallData;
	
	    static createFrom(source: any = {}) {
	        return new Card(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.quantity = source["quantity"];
	        this.name = source["name"];
	        this.setCode = source["setCode"];
	        this.collectorNumber = source["collectorNumber"];
	        this.foil = source["foil"];
	        this.tags = source["tags"];
	        this.scryFall = this.convertValues(source["scryFall"], ScryfallData);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DeckInfo {
	    title: string;
	    status: string;
	    colors: string;
	    commander: string;
	    strategy: string;
	    universe?: string;
	
	    static createFrom(source: any = {}) {
	        return new DeckInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.status = source["status"];
	        this.colors = source["colors"];
	        this.commander = source["commander"];
	        this.strategy = source["strategy"];
	        this.universe = source["universe"];
	    }
	}
	export class Deck {
	    slug: string;
	    info: DeckInfo;
	    cards: Card[];
	    wishlist: Card[];
	    cardCount: number;
	
	    static createFrom(source: any = {}) {
	        return new Deck(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.slug = source["slug"];
	        this.info = this.convertValues(source["info"], DeckInfo);
	        this.cards = this.convertValues(source["cards"], Card);
	        this.wishlist = this.convertValues(source["wishlist"], Card);
	        this.cardCount = source["cardCount"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

export namespace main {
	
	export class CollectionInfo {
	    path: string;
	    label: string;
	    lastOpened: string;
	    isActive: boolean;
	    isValid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CollectionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.label = source["label"];
	        this.lastOpened = source["lastOpened"];
	        this.isActive = source["isActive"];
	        this.isValid = source["isValid"];
	    }
	}
	export class AppState {
	    hasCollection: boolean;
	    collectionPath: string;
	    collectionLabel: string;
	    collectionValid: boolean;
	    collections: CollectionInfo[];
	    needsSetup: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hasCollection = source["hasCollection"];
	        this.collectionPath = source["collectionPath"];
	        this.collectionLabel = source["collectionLabel"];
	        this.collectionValid = source["collectionValid"];
	        this.collections = this.convertValues(source["collections"], CollectionInfo);
	        this.needsSetup = source["needsSetup"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class DeckSummary {
	    slug: string;
	    title: string;
	    commander: string;
	    colors: string;
	    status: string;
	    cardCount: number;
	    universe?: string;
	
	    static createFrom(source: any = {}) {
	        return new DeckSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.slug = source["slug"];
	        this.title = source["title"];
	        this.commander = source["commander"];
	        this.colors = source["colors"];
	        this.status = source["status"];
	        this.cardCount = source["cardCount"];
	        this.universe = source["universe"];
	    }
	}

}

