import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { catchError, map, tap } from 'rxjs/operators';
import { environment } from 'src/environments/environment';
import { List } from './common';
import { Logger, LoggingService } from './logging.service';
import { Source } from './source.service';

@Injectable({
  providedIn: 'root'
})
export class StationService {

  constructor(private http: HttpClient,
    logging: LoggingService) {
    this.logger = logging.get('StationService');
  }

  private logger: Logger;

  list(): Observable<Station[]> {
    const url = `${environment.apiURL}stations`;
    this.logger.log(`Retrieving stations`, url);
    return this.http.get<List<Station>>(url)
      .pipe(
        tap(res => this.logger.log('Retrieved stations', res)),
        catchError(this.logger.handleError<List<Station>>(`list()`)),
        map(res => res ? res.items : [])
      );
  }

  get(name: string): Observable<Station> {
    const uriName = encodeURIComponent(name);
    const url = `${environment.apiURL}stations/${uriName}`;
    this.logger.log(`Retrieving station details for ${name}`, url);
    return this.http.get<Station>(url)
      .pipe(
        tap(res => this.logger.log('Retrieved station details', res)),
        catchError(this.logger.handleError<Station>(`get('${name}')`))
      );
  }
}

export interface Station {
  name: string;
  status: string;
  sources: Source[];
}
