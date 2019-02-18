import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { catchError, map, tap } from 'rxjs/operators';
import { environment } from 'src/environments/environment';
import { List } from './common';
import { Logger, LoggingService } from './logging.service';
import { Source } from './source.service';
import { Result, Results } from './Result';
import { ValueList } from './data';

@Injectable({
  providedIn: 'root'
})
export class StationService {

  constructor(private http: HttpClient,
    logging: LoggingService) {
    this.logger = logging.get('StationService');
  }

  private logger: Logger;

  list(): Observable<Results<Station>> {
    const url = `${environment.apiURL}stations`;
    this.logger.log(`Retrieving stations`, url);
    return this.http.get<List<Station>>(url)
      .pipe(
        tap(res => this.logger.log('Retrieved stations', res)),
        map(res => Results.new(res ? res.items : [])),
        catchError(this.logger.handleError(`list()`, Results.new<Station>(null, 'Unable to load stations')))
      );
  }

  get(name: string): Observable<Result<Station>> {
    const uriName = encodeURIComponent(name),
      url = `${environment.apiURL}stations/${uriName}`;
    this.logger.log(`Retrieving station details for ${name}`, url);
    return this.http.get<Station>(url)
      .pipe(
        tap(res => this.logger.log('Retrieved station details', res)),
        map(res => Result.new(res)),
        catchError(this.logger.handleError(`get('${name}')`, Result.new<Station>(null, 'Unable to load station')))
      );
  }

  getValues(station: string, source: string): Observable<Result<ValueList>> {
    const uriStation = encodeURIComponent(station),
      uriSource = encodeURIComponent(source),
    url = `${environment.apiURL}stations/${uriStation}/sources/${uriSource}/values`;
    this.logger.log(`Retrieving source ${source} values from station ${name}`, url);
    return this.http.get<ValueList>(url)
      .pipe(
        tap(res => this.logger.log('Retrieved source values from station', res)),
        map(res => Result.new(res)),
        catchError(this.logger.handleError(`getValues('${station}', '${source})`, Result.new<ValueList>(null, 'Unable to load source values from station')))
      );
  }
}

export interface Station {
  name: string;
  status: string;
  sources: Source[];
}
