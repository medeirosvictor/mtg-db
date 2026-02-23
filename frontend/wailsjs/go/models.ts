export namespace deck {
	
	export class Card {
	    quantity: number;
	    name: string;
	    setCode?: string;
	    collectorNumber?: string;
	    foil?: boolean;
	
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

