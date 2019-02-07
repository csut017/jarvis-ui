import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { LoggingService, Logger } from './logging.service';
import { Observable } from 'rxjs';
import { environment } from 'src/environments/environment';
import { tap, catchError, map } from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class StationService {

  constructor(private http: HttpClient,
    private logging: LoggingService) {
    this.logger = logging.get('StationService');
    this.current = null;
  }

  current: StationInterface;
  private logger: Logger;

  list(): Observable<StationInterface[]> {
    const url = `${environment.apiURL}stations`;
    this.logger.log(`Retrieving stations`, url);
    return this.http.get<stationList>(url)
      .pipe(
        tap(res => this.logger.log('Retrieved stations', res)),
        map(res => res.items.map(item => new StationInterface(this.http, this.logging, item.name, item.status))),
        catchError(this.logger.handleError<StationInterface[]>(`list()`))
      );
  }
}

export class StationInterface {
  constructor(private http: HttpClient,
    logging: LoggingService,
    public name: string,
    public status: string) {
    this.logger = logging.get('StationInterface');
    this.uriName = encodeURIComponent(this.name);
  }

  private uriName: string;
  private logger: Logger;

  get(): Observable<StationDetails> {
    const url = `${environment.apiURL}stations/${this.uriName}`;
    this.logger.log(`Retrieving station details for ${this.name}`, url);
    return this.http.get<StationDetails>(url)
      .pipe(
        tap(res => this.logger.log('Retrieved station details', res)),
        catchError(this.logger.handleError<StationDetails>(`'${this.name}'->get()`))
      );
  }
}

export interface StationDetails {
}

interface stationListItem {
  name: string;
  status: string;
}

interface stationList {
  items: stationListItem[];
}